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

func getChainCerSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_cer c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.sn and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainCerSqlsDb(): lab_rpki_cer id fail:", err)
		return nil, err
	}
	belogs.Info("getChainCerSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Now().Sub(start).Seconds())
	return chainCertSqls, nil
}

func updateCersDb(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	cerIds := chains.CerIds
	for _, cerId := range cerIds {
		err = updateCerDb(session, chains, cerId)
		if err != nil {
			belogs.Error("updateCersDb(): updateCerDb fail :", cerId, err)
			xormdb.RollbackAndLogError(session, "updateCersDb(): updateCerDb fail: "+
				convert.ToString(cerId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateCersDb(): CommitSession fail :", err)
		return
	}
	belogs.Debug("updateCersDb(): len(cerIds):", len(cerIds), "  time(s):", time.Now().Sub(start).Seconds())
	return
}

func updateCerDb(session *xorm.Session, chains *Chains, cerId uint64) (err error) {
	start := time.Now()
	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("updateCerDb(): GetCer fail :", cerId, err)
		return err
	}

	chainDbCerModel := NewChainDbCerModel(&chainCer)
	originModel := model.JudgeOrigin(chainCer.FilePath)
	belogs.Debug("updateCerDb():chainDbCerModel, id, len(chainDbCerModel.ChildChainCers):", chainDbCerModel.Id,
		len(chainDbCerModel.ChildChainCers), originModel)

	chainCerts := jsonutil.MarshalJson(*chainDbCerModel)
	state := jsonutil.MarshalJson(chainCer.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateCerDb():cerId:", cerId, "    chainCerts", chainCerts, "   state:", jsonutil.MarshalJson(state))
	sqlStr := `UPDATE lab_rpki_cer set chainCerts=?, state=? , origin=?  where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, cerId)
	if err != nil {
		belogs.Error("updateCerDb(): UPDATE lab_rpki_cer fail :", cerId, err)
		return err
	}
	belogs.Debug("updateCerDb(): cerId:", cerId, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
