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

func GetChainMftIds() (mftIds []uint64, err error) {
	start := time.Now()
	err = xormdb.XormEngine.Table("lab_rpki_mft").Cols("id").Find(&mftIds)
	if err != nil {
		belogs.Error("GetChainCrlIds(): lab_rpki_mft id fail:", err)
		return nil, err
	}
	belogs.Debug("GetChainMftIds(): len(mftIds):", len(mftIds), "  time(s):", time.Now().Sub(start).Seconds())
	return mftIds, nil
}

func GetChainMft(mftId uint64) (chainMft chainmodel.ChainMft, err error) {
	start := time.Now()
	chainMftSql := chainmodel.ChainMftSql{}
	belogs.Debug("GetChainMft(): mftId:", mftId)
	has, err := xormdb.XormEngine.Table("lab_rpki_mft").
		Select("id,ski,aki,filePath,fileName,mftNumber,state,jsonAll->'$.eeCertModel.eeCertStart' as eeCertStart,jsonAll->'$.eeCertModel.eeCertEnd' as eeCertEnd").
		Where("id=?", mftId).Get(&chainMftSql)
	if err != nil {
		belogs.Error("GetChainMft(): lab_rpki_mft fail:", mftId, err)
		return chainMft, err
	}
	belogs.Debug("GetChainMft(): mftId:", mftId, "  has:", has)
	chainMft = chainMftSql.ToChainMft()
	belogs.Debug("GetChainMft(): mftId:", mftId, "    chainMft.Id:", chainMft.Id)

	// get current stateModel
	chainMft.StateModel = model.GetStateModelAndResetStage(chainMft.State, "chainvalidate")

	// get filehash
	chainMft.ChainFileHashs, err = getChainFileHashs(chainMft.Id, chainMft.Aki)
	if err != nil {
		belogs.Error("GetChainMft(): getChainFileHashs fail, chainMft.Id:", chainMft.Id, err)
		return chainMft, err
	}

	// get sn in crl
	chainMft.ChainSnInCrlRevoked, err = getMftEeSnInCrlRevoked(mftId)
	if err != nil {
		belogs.Error("GetChainMft(): getMftEeSnInCrlRevoked fail, chainMft.Id:", chainMft.Id, err)
		return chainMft, err
	}
	belogs.Debug("GetChainMft(): mftId:", mftId, "    chainMft.Id:", chainMft.Id, "  time(s):", time.Now().Sub(start).Seconds())
	return chainMft, nil
}

func getChainFileHashs(mftId uint64, aki string) (chainFileHashs []chainmodel.ChainFileHash, err error) {
	start := time.Now()
	err = xormdb.XormEngine.Table("lab_rpki_mft_file_hash").
		Cols("file,hash").
		Where("mftId=?", mftId).
		OrderBy("id").Find(&chainFileHashs)
	if err != nil {
		belogs.Error("getChainFileHashs(): lab_rpki_mft_file_hash fail:", err)
		return chainFileHashs, err
	}

	for i := range chainFileHashs {
		filePath, err := getPath(chainFileHashs[i].File, aki)
		if err != nil {
			belogs.Error("getChainFileHashs(): getPath fail:", chainFileHashs[i].File, aki)
			return chainFileHashs, err
		}
		chainFileHashs[i].Path = filePath
	}

	belogs.Debug("getChainFileHashs():mftId:",
		mftId, "    len(chainFileHashs)", len(chainFileHashs), "  time(s):", time.Now().Sub(start).Seconds())
	return chainFileHashs, nil
}

func getPath(fileName, aki string) (filePath string, err error) {
	start := time.Now()

	sqls := " select filePath from lab_rpki_cer where fileName='" + fileName + "' and aki='" + aki + "' " +
		" union " +
		" select filePath from lab_rpki_crl where fileName='" + fileName + "' and aki='" + aki + "' " +
		" union " +
		" select filePath from lab_rpki_mft where fileName='" + fileName + "' and aki='" + aki + "' " +
		" union " +
		" select filePath from lab_rpki_roa where fileName='" + fileName + "' and aki='" + aki + "' "
	belogs.Debug("GetFileHash(): select fileName and aki:", fileName, aki)

	_, err = xormdb.XormEngine.SQL(sqls).Get(&filePath)
	if err != nil {
		belogs.Error("getPath(): select fileName and aki fail :", sqls, err)
		return "", err
	}
	belogs.Debug("getPath(): fileName, aki:", fileName, aki, "  time(s):", time.Now().Sub(start).Seconds())
	return filePath, nil
}

func getMftEeSnInCrlRevoked(mftId uint64) (chainSnInCrlRevoked chainmodel.ChainSnInCrlRevoked, err error) {
	start := time.Now()
	sql := `select l.fileName, r.revocationTime from lab_rpki_mft c, lab_rpki_crl l, lab_rpki_crl_revoked_cert r
	 where  c.jsonAll->'$.eeCertModel.sn' = r.sn and r.crlId = l.id and c.aki = l.aki and c.id=` + convert.ToString(mftId)
	belogs.Debug("getMftEeSnInCrlRevoked(): mftId:", mftId, "   sql:", sql)
	_, err = xormdb.XormEngine.
		Sql(sql).Get(&chainSnInCrlRevoked)
	if err != nil {
		belogs.Error("getMftEeSnInCrlRevoked(): select fail:", mftId, err)
		return chainSnInCrlRevoked, err
	}
	belogs.Debug("getMftEeSnInCrlRevoked(): mftId:", mftId, "  time(s):", time.Now().Sub(start).Seconds())
	return chainSnInCrlRevoked, nil

}

func UpdateMfts(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
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
