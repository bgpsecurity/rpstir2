package db

import (
	"bytes"
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

func GetChainCrlIds() (crlIds []uint64, err error) {
	start := time.Now()
	err = xormdb.XormEngine.Table("lab_rpki_crl").Cols("id").Find(&crlIds)
	if err != nil {
		belogs.Error("GetChainCrlIds(): lab_rpki_crl id fail:", err)
		return nil, err
	}
	belogs.Debug("GetChainCrlIds(): len(crlIds):", len(crlIds), "  time(s):", time.Now().Sub(start).Seconds())
	return crlIds, nil
}

func GetChainCrl(crlId uint64) (chainCrl chainmodel.ChainCrl, err error) {
	start := time.Now()

	chainCrlSql := chainmodel.ChainCrlSql{}
	has, err := xormdb.XormEngine.Table("lab_rpki_crl").
		Cols("id,aki,filePath,fileName,state").
		Where("id=?", crlId).Get(&chainCrlSql)
	if err != nil {
		belogs.Error("GetChainCrl(): lab_rpki_crl fail:", crlId, err)
		return chainCrl, err
	}
	belogs.Debug("GetChainCrl(): crlId:", crlId, "  has:", has)
	chainCrl = chainCrlSql.ToChainCrl()
	belogs.Debug("GetChainCrl(): crlId:", crlId, "  chainCrl.Id:", chainCrl.Id)

	// get current stateModel
	chainCrl.StateModel = model.GetStateModelAndResetStage(chainCrl.State, "chainvalidate")

	// get revoked certs
	chainCrl.ChainRevokedCerts, err = getChainRevokedCerts(chainCrl.Id)
	if err != nil {
		belogs.Error("GetChainCrl(): getChainRevokedCerts fail, chainCrl.Id:", chainCrl.Id, err)
		return chainCrl, err
	}

	chainCrl.ShouldRevokedCerts, err = getRevokedCerts(&chainCrl)
	if err != nil {
		belogs.Error("GetChainCrl(): getRevokedCerts fail, chainCrl.Id:", chainCrl.Id, err)
		return chainCrl, err
	}

	belogs.Debug("GetChainCrl(): crlId:", crlId, "    chainCrl.Id:", chainCrl.Id, "  time(s):", time.Now().Sub(start).Seconds())
	return chainCrl, nil
}
func getChainRevokedCerts(crlId uint64) (chainRevokedCerts []chainmodel.ChainRevokedCert, err error) {
	start := time.Now()

	err = xormdb.XormEngine.Table("lab_rpki_crl_revoked_cert").
		Cols("sn").
		Where("crlId=?", crlId).
		OrderBy("id").Find(&chainRevokedCerts)
	if err != nil {
		belogs.Error("getChainRevokedCerts(): lab_rpki_crl_revoked_cert fail, crlId:", crlId, err)
		return chainRevokedCerts, err
	}
	belogs.Debug("getChainRevokedCerts():crlId, len(chainFileHashs):", crlId, len(chainRevokedCerts), "  time(s):", time.Now().Sub(start).Seconds())
	return chainRevokedCerts, nil
}

func getRevokedCerts(chainCrl *chainmodel.ChainCrl) (shouldRevokedCerts []string, err error) {
	start := time.Now()
	if len(chainCrl.ChainRevokedCerts) == 0 {
		return
	}

	shouldRevokedCerts = make([]string, 0)

	var b bytes.Buffer
	for i, c := range chainCrl.ChainRevokedCerts {
		if i < len(chainCrl.ChainRevokedCerts)-1 {
			b.WriteString("'" + c.Sn + "',")
		} else {
			b.WriteString("'" + c.Sn + "'")
		}
	}
	belogs.Debug("getRevokedCerts():revokedcerts sn :", b.String())

	//get same sn in cer
	sql := `select CONCAT(c.filePath,c.fileName) from lab_rpki_cer c  
	 where c.sn in (` + b.String() + `) and c.aki = '` + chainCrl.Aki + `' order by c.id`
	belogs.Debug("getRevokedCerts():select lab_rpki_cer, sql:", sql)
	chainCers := make([]string, 0)
	err = xormdb.XormEngine.
		Sql(sql).
		Find(&chainCers)
	if err != nil {
		belogs.Error("getRevokedCerts(): select lab_rpki_cer fail :", sql, err)
		return nil, err
	}
	if len(chainCers) > 0 {
		shouldRevokedCerts = append(shouldRevokedCerts, chainCers...)
		belogs.Debug("getRevokedCerts(): shouldRevokedCerts append chainCers:", shouldRevokedCerts)
	}

	// get same sn in roa
	sql = `select CONCAT(c.filePath,c.fileName) from lab_rpki_roa c  
	 where c.jsonAll->'$.eeCertModel.sn' in (` + b.String() + `) and c.aki = '` + chainCrl.Aki + `' order by c.id`
	belogs.Debug("getRevokedCerts():select lab_rpki_roa, sql:", sql)
	chaiRoas := make([]string, 0)
	err = xormdb.XormEngine.
		Sql(sql).
		Find(&chaiRoas)
	if err != nil {
		belogs.Error("getRevokedCerts(): select  lab_rpki_roa fail :", sql, err)
		return nil, err
	}
	if len(chaiRoas) > 0 {
		shouldRevokedCerts = append(shouldRevokedCerts, chaiRoas...)
		belogs.Debug("getRevokedCerts(): shouldRevokedCerts append chaiRoas:", shouldRevokedCerts)
	}

	// get same sn in mft
	sql = `select CONCAT(c.filePath,c.fileName) from lab_rpki_mft c  
	 where c.jsonAll->'$.eeCertModel.sn' in (` + b.String() + `) and c.aki = '` + chainCrl.Aki + `' order by c.id`
	belogs.Debug("getRevokedCerts():select lab_rpki_mft, sql:", sql)
	chaiMfts := make([]string, 0)
	err = xormdb.XormEngine.
		Sql(sql).
		Find(&chaiMfts)
	if err != nil {
		belogs.Error("getRevokedCerts(): select  lab_rpki_mft fail :", sql, err)
		return nil, err
	}
	if len(chaiRoas) > 0 {
		shouldRevokedCerts = append(shouldRevokedCerts, chaiMfts...)
		belogs.Debug("getRevokedCerts(): shouldRevokedCerts append chaiMfts:", shouldRevokedCerts)
	}

	belogs.Debug("getRevokedCerts(): chainCrl.Id:", chainCrl.Id, "  time(s):", time.Now().Sub(start).Seconds())
	return shouldRevokedCerts, nil
}

func UpdateCrls(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	crlIds := chains.CrlIds
	for _, crlId := range crlIds {
		err = updateCrl(session, chains, crlId)
		if err != nil {
			belogs.Error("UpdateCrls(): updateCrl fail :", crlId, err)
			xormdb.RollbackAndLogError(session, "UpdateCrls(): updateCrl fail: "+convert.ToString(crlId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateCrls(): CommitSession fail :", err)
		return
	}
	belogs.Debug("UpdateCrls():len(crlIds):", len(crlIds), "  time(s):", time.Now().Sub(start).Seconds())

}

func updateCrl(session *xorm.Session, chains *chainmodel.Chains, crlId uint64) (err error) {

	start := time.Now()
	chainCrl, err := chains.GetCrlById(crlId)
	if err != nil {
		belogs.Error("updateCrl(): GetCrl fail :", crlId, err)
		return err
	}

	chainDbCrlModel := chainmodel.NewChainDbCrlModel(&chainCrl)
	originModel := model.JudgeOrigin(chainCrl.FilePath)

	chainCerts := jsonutil.MarshalJson(*chainDbCrlModel)
	state := jsonutil.MarshalJson(chainCrl.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateCrl():crlId:", crlId, "   chainCrl:", jsonutil.MarshalJson(chainCrl),
		"   chainDbCrlModel chainCerts:", chainCerts, "   origin:", origin, "  state:", state)
	sqlStr := `UPDATE lab_rpki_crl set chainCerts=?, state=?, origin=?   where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, crlId)
	if err != nil {
		belogs.Error("updateCrl(): UPDATE lab_rpki_crl fail :", crlId, err)
		return err
	}
	belogs.Debug("updateCrl(): crlId:", crlId, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
