package chainvalidate

import (
	"sync"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
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
	belogs.Info("getChainMftSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Now().Sub(start).Seconds())
	return chainCertSqls, nil
}

func GetChainFileHashs(mftId uint64) (chainFileHashs []ChainFileHash, err error) {
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

	belogs.Debug("GetChainFileHashs():mftId:",
		mftId, "    len(chainFileHashs)", len(chainFileHashs), "  time(s):", time.Now().Sub(start).Seconds())
	return chainFileHashs, nil
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
	belogs.Debug("updateMftsDb(): len(mftIds):", len(mftIds), "  time(s):", time.Now().Sub(start).Seconds())

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
	belogs.Debug("updateMftDb():mftId:", mftId, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
