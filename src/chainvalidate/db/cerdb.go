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

func GetChainCerIds() (cerIds []uint64, err error) {
	start := time.Now()
	err = xormdb.XormEngine.Table("lab_rpki_cer").Cols("id").Find(&cerIds)
	if err != nil {
		belogs.Error("GetChainCerIds(): lab_rpki_cer id fail:", err)
		return nil, err
	}
	belogs.Debug("GetChainCerIds(): len(cerIds):", len(cerIds), "  time(s):", time.Now().Sub(start).Seconds())
	return cerIds, nil
}

func GetChainCer(cerId uint64) (chainCer chainmodel.ChainCer, err error) {
	start := time.Now()
	belogs.Debug("GetChainCer(): cerId:", cerId)

	chainCerSql := chainmodel.ChainCerSql{}
	has, err := xormdb.XormEngine.Table("lab_rpki_cer").
		Select("id,ski,aki,filePath,fileName,state,jsonAll->'$.isRoot' as isRoot").
		Where("id=?", cerId).Get(&chainCerSql)
	if err != nil {
		belogs.Error("GetChainCer(): lab_rpki_cer fail:", cerId, err)
		return chainCer, err
	}
	belogs.Debug("GetChainCer(): cerId:", cerId, "  has:", has)
	chainCer = chainCerSql.ToChainCer()
	belogs.Debug("GetChainCer(): cerId:", cerId, "  chainCer.Id:", chainCer.Id)

	// get current stateModel
	chainCer.StateModel = model.GetStateModelAndResetStage(chainCer.State, "chainvalidate")

	// get ipaddress

	//lab_rpki_cer_ipaddress
	belogs.Debug("GetChainCer():chainCer.id:", chainCer.Id)
	chainCer.ChainIpAddresses, err = getChainCerIpAddresses(cerId)
	if err != nil {
		belogs.Error("GetChainCer(): getChainIpAddresses fail, chainCer.Id:", chainCer.Id, err)
		return chainCer, err
	}

	// lab_rpki_cer_asn
	belogs.Debug("GetChainCer():chainCer.id:", chainCer.Id)
	chainCer.ChainAsns, err = getChainAsns(cerId)
	if err != nil {
		belogs.Error("GetChainCer(): getChainAsns fail, chainCer.Id:", chainCer.Id, err)
		return chainCer, err
	}

	chainCer.ChainSnInCrlRevoked, err = getCerSnInCrlRevoked(cerId)
	if err != nil {
		belogs.Error("GetChainCer(): getSnInCrlRevoked fail, chainCer.Id:", chainCer.Id, err)
		return chainCer, err
	}

	belogs.Debug("GetChainCer(): cerId:", cerId, "   chainCer.Id:", chainCer.Id, "  time(s):", time.Now().Sub(start).Seconds())
	return chainCer, nil
}

func getChainCerIpAddresses(cerId uint64) (chainIpAddresses []chainmodel.ChainIpAddress, err error) {
	start := time.Now()

	belogs.Debug("getChainCerIpAddresses(): cerId:", cerId)
	chainIpAddresses = make([]chainmodel.ChainIpAddress, 0)
	err = xormdb.XormEngine.Table("lab_rpki_cer_ipaddress").
		Select("id,rangeStart,rangeEnd").
		Where("cerId=?", cerId).
		OrderBy("id").Find(&chainIpAddresses)
	if err != nil {
		belogs.Error("getChainCerIpAddresses(): lab_rpki_cer_ipaddress fail:", err)
		return chainIpAddresses, err
	}
	if len(chainIpAddresses) < 3 {
		belogs.Debug("getChainCerIpAddresses():cerId, chainIpAddresses:", cerId, jsonutil.MarshalJson(chainIpAddresses))
	} else {
		belogs.Debug("getChainCerIpAddresses():cerId, len(chainIpAddresses):", cerId, len(chainIpAddresses))
	}
	belogs.Debug("getChainCerIpAddresses(): cerId:", cerId, "  time(s):", time.Now().Sub(start).Seconds())
	return chainIpAddresses, nil
}

func getChainAsns(cerId uint64) (chainAsns []chainmodel.ChainAsn, err error) {
	start := time.Now()

	belogs.Debug("getChainAsns(): cerId:", cerId)
	chainAsns = make([]chainmodel.ChainAsn, 0)
	err = xormdb.XormEngine.Table("lab_rpki_cer_asn").
		Select("id,asn,min,max").
		Where("cerId=?", cerId).
		OrderBy("id").Find(&chainAsns)
	if err != nil {
		belogs.Error("getChainAsns(): lab_rpki_cer_asn fail:", err)
		return chainAsns, err
	}
	if len(chainAsns) < 3 {
		belogs.Debug("getChainAsns():cerId, chainAsns:",
			cerId, jsonutil.MarshalJson(chainAsns))
	} else {
		belogs.Debug("getChainAsns():cerId, len(chainAsns):",
			cerId, len(chainAsns))
	}
	belogs.Debug("getChainAsns(): cerId:", cerId, "  time(s):", time.Now().Sub(start).Seconds())
	return chainAsns, nil
}

func getCerSnInCrlRevoked(cerId uint64) (chainSnInCrlRevoked chainmodel.ChainSnInCrlRevoked, err error) {
	start := time.Now()
	sql := `select l.fileName, r.revocationTime from lab_rpki_cer c, lab_rpki_crl l, lab_rpki_crl_revoked_cert r
	 where  c.sn = r.sn and r.crlId = l.id and c.aki = l.aki and c.id=` + convert.ToString(cerId)
	belogs.Debug("getCerSnInCrlRevoked(): cerId:", cerId, "   sql:", sql)
	_, err = xormdb.XormEngine.
		Sql(sql).Get(&chainSnInCrlRevoked)
	if err != nil {
		belogs.Error("getCerSnInCrlRevoked(): select fail:", cerId, err)
		return chainSnInCrlRevoked, err
	}
	belogs.Debug("getCerSnInCrlRevoked(): cerId:", cerId, "  time(s):", time.Now().Sub(start).Seconds())
	return chainSnInCrlRevoked, nil

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
