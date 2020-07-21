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

func GetChainCerSqls() (chainCertSqls []chainmodel.ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]chainmodel.ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_cer c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.sn and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("GetChainCerSqls(): lab_rpki_cer id fail:", err)
		return nil, err
	}
	belogs.Info("GetChainCerSqls(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Now().Sub(start).Seconds())
	return chainCertSqls, nil
}

func UpdateCers(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	cerIds := chains.CerIds
	for _, cerId := range cerIds {
		err = updateCer(session, chains, cerId)
		if err != nil {
			belogs.Error("UpdateCers(): updateCer fail :", cerId, err)
			xormdb.RollbackAndLogError(session, "UpdateCers(): updateCer fail: "+
				convert.ToString(cerId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateCers(): CommitSession fail :", err)
		return
	}
	belogs.Debug("UpdateCers(): len(cerIds):", len(cerIds), "  time(s):", time.Now().Sub(start).Seconds())
	return
}

func updateCer(session *xorm.Session, chains *chainmodel.Chains, cerId uint64) (err error) {
	start := time.Now()
	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("updateCer(): GetCer fail :", cerId, err)
		return err
	}

	chainDbCerModel := chainmodel.NewChainDbCerModel(&chainCer)
	originModel := model.JudgeOrigin(chainCer.FilePath)
	belogs.Debug("updateCer():chainDbCerModel, id, len(chainDbCerModel.ChildChainCers):", chainDbCerModel.Id,
		len(chainDbCerModel.ChildChainCers), originModel)

	chainCerts := jsonutil.MarshalJson(*chainDbCerModel)
	state := jsonutil.MarshalJson(chainCer.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateCer():cerId:", cerId, "    chainCerts", chainCerts, "   state:", jsonutil.MarshalJson(state))
	sqlStr := `UPDATE lab_rpki_cer set chainCerts=?, state=? , origin=?  where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, cerId)
	if err != nil {
		belogs.Error("updateCer(): UPDATE lab_rpki_cer fail :", cerId, err)
		return err
	}
	belogs.Debug("updateCer(): cerId:", cerId, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
