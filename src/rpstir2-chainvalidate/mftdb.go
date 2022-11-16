package chainvalidate

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

func getChainMftSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_mft c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainMftSqlsDb(): lab_rpki_mft id fail:", err)
		return nil, err
	}
	belogs.Info("getChainMftSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func GetChainFileHashsDb(mftId uint64) (chainFileHashs []ChainFileHash, err error) {
	start := time.Now()
	sql := `select  v.file as file, v.hash as hash,CONCAT(IFNULL(cer.filePath,''),IFNULL(crl.filePath,''),IFNULL(roa.filePath,'')) as path 
			from lab_rpki_mft_file_hash_view v 
				left join lab_rpki_cer cer on cer.aki=v.aki and cer.fileName=v.file 
				left join lab_rpki_crl crl on crl.aki=v.aki and crl.fileName=v.file 
				left join lab_rpki_roa roa on roa.aki=v.aki and roa.fileName=v.file 
			where v.mftId=?  
			order by v.mftFileHashId `
	err = xormdb.XormEngine.SQL(sql, mftId).Find(&chainFileHashs)
	if err != nil {
		belogs.Error("getChainFileHashs(): lab_rpki_mft_file_hash fail:", err)
		return chainFileHashs, err
	}

	belogs.Debug("GetChainFileHashsDb():mftId:",
		mftId, "    len(chainFileHashs)", len(chainFileHashs), "  time(s):", time.Since(start))
	return chainFileHashs, nil
}

func GetPreviousMftDb(mftId uint64) (previousMft PreviousMft, err error) {
	start := time.Now()
	/* // because using json directly, it will cause ' Out of sort memory, consider increasing server sort buffer size'
	sql := `select f.jsonAll->>'$.mftNumber' as mftNumber,f.jsonAll->>'$.thisUpdate' as thisUpdate, f.jsonAll->>'$.nextUpdate' as nextUpdate
	        from  lab_rpki_sync_log_file f ,
				( select m.mftNumber,m.filePath,m.fileName,m.syncLogId from lab_rpki_mft m where m.id = ? ) t
			where f.filePath = t.filePath and f.fileName=t.fileName and f.syncLogId < t.syncLogId
			order by f.syncLogId desc limit 1  `
	*/
	// get id ,then select from json
	sql := `select  ff.jsonAll->>'$.mftNumber' as mftNumber,ff.jsonAll->>'$.thisUpdate' as thisUpdate, ff.jsonAll->>'$.nextUpdate' as nextUpdate 
			from lab_rpki_sync_log_file ff where ff.id = ( 
				select f.id from  lab_rpki_sync_log_file f , ( select m.mftNumber,m.filePath,m.fileName,m.syncLogId from lab_rpki_mft m where m.id = ? ) t 
				where f.filePath = t.filePath and f.fileName=t.fileName and f.syncLogId < t.syncLogId 
                order by f.syncLogId desc limit 1 
	     	)`
	found, err := xormdb.XormEngine.SQL(sql, mftId).Get(&previousMft)
	if err != nil {
		belogs.Error("GetPreviousMftDb(): lab_rpki_sync_log_file fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return previousMft, err
	}
	previousMft.Found = found
	belogs.Info("GetPreviousMftDb():mftId:", mftId, "   previousMft:", previousMft,
		"  time(s):", time.Since(start)) //shaodebug
	return previousMft, nil
}

func updateMftsDb(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	mftIds := chains.MftIds
	for _, mftId := range mftIds {

		err = updateMftDb(session, chains, mftId)
		if err != nil {
			belogs.Error("updateMftsDb(): updateMftDb fail, mftId:", mftId, err)
			xormdb.RollbackAndLogError(session, "updateMftsDb(): updateMftDb fail: "+convert.ToString(mftId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateMftsDb(): CommitSession fail :", err)
		return
	}
	belogs.Debug("updateMftsDb(): len(mftIds):", len(mftIds), "  time(s):", time.Since(start))

}

func updateMftDb(session *xorm.Session, chains *Chains, mftId uint64) (err error) {
	start := time.Now()
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("updateMftDb(): GetMft fail :", mftId, err)
		return err
	}

	chainDbMftModel := NewChainDbMftModel(&chainMft)
	originModel := model.JudgeOrigin(chainMft.FilePath)

	chainCerts := jsonutil.MarshalJson(*chainDbMftModel)
	state := jsonutil.MarshalJson(chainMft.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateMftDb():mftId:", mftId, "   chainCerts:", chainCerts, "  origin:", origin, " state:", chainCerts, state)
	sqlStr := `UPDATE lab_rpki_mft set chainCerts=?, state=?, origin=?   where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, mftId)
	if err != nil {
		belogs.Error("updateMftDb(): UPDATE lab_rpki_mft fail :", mftId, err)
		return err
	}
	belogs.Debug("updateMftDb():mftId:", mftId, "  time(s):", time.Since(start))
	return nil
}
