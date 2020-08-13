package db

import (
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	chainmodel "chainvalidate/model"
	"model"
)

func GetChainMftSqls() (chainCertSqls []chainmodel.ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]chainmodel.ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_mft c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("GetChainMftSqls(): lab_rpki_mft id fail:", err)
		return nil, err
	}
	belogs.Info("GetChainMftSqls(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Now().Sub(start).Seconds())
	return chainCertSqls, nil
}

func GetChainFileHashs(mftId uint64) (chainFileHashs []chainmodel.ChainFileHash, err error) {
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

func UpdateMfts(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	mftIds := chains.MftIds
	for _, mftId := range mftIds {

		err = updateMft(session, chains, mftId)
		if err != nil {
			belogs.Error("UpdateMfts(): updateMft fail, mftId:", mftId, err)
			xormdb.RollbackAndLogError(session, "UpdateMfts(): updateMft fail: "+convert.ToString(mftId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateMfts(): CommitSession fail :", err)
		return
	}
	belogs.Debug("UpdateMfts(): len(mftIds):", len(mftIds), "  time(s):", time.Now().Sub(start).Seconds())

}

func updateMft(session *xorm.Session, chains *chainmodel.Chains, mftId uint64) (err error) {
	start := time.Now()
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("updateMft(): GetMft fail :", mftId, err)
		return err
	}

	chainDbMftModel := chainmodel.NewChainDbMftModel(&chainMft)
	originModel := model.JudgeOrigin(chainMft.FilePath)

	chainCerts := jsonutil.MarshalJson(*chainDbMftModel)
	state := jsonutil.MarshalJson(chainMft.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateMft():mftId:", mftId, "   chainCerts:", chainCerts, "  origin:", origin, " state:", chainCerts, state)
	sqlStr := `UPDATE lab_rpki_mft set chainCerts=?, state=?, origin=?   where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, mftId)
	if err != nil {
		belogs.Error("updateMft(): UPDATE lab_rpki_mft fail :", mftId, err)
		return err
	}
	belogs.Debug("updateMft():mftId:", mftId, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
