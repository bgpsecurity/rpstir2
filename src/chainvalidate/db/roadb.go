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

func GetChainRoaIds() (roaIds []uint64, err error) {
	start := time.Now()
	err = xormdb.XormEngine.Table("lab_rpki_roa").Cols("id").Find(&roaIds)
	if err != nil {
		belogs.Error("GetChainRoaIds(): lab_rpki_roa id fail:", err)
		return nil, err
	}
	belogs.Debug("GetChainRoaIds(): len(roaIds):", len(roaIds), "  time(s):", time.Now().Sub(start).Seconds())
	return roaIds, nil
}

func GetChainRoa(roaId uint64) (chainRoa chainmodel.ChainRoa, err error) {
	start := time.Now()

	chainRoaSql := chainmodel.ChainRoaSql{}
	has, err := xormdb.XormEngine.Table("lab_rpki_roa").
		Select("id,asn,ski,aki,filePath,fileName,state,jsonAll->'$.eeCertModel.eeCertStart' as eeCertStart,jsonAll->'$.eeCertModel.eeCertEnd' as eeCertEnd").
		Where("id=?", roaId).Get(&chainRoaSql)
	if err != nil {
		belogs.Error("GetChainRoa(): lab_rpki_roa fail:", err)
		return chainRoa, err
	}
	belogs.Debug("GetChainRoa(): roaId:", roaId, "  has:", has)
	chainRoa = chainRoaSql.ToChainRoa()
	belogs.Debug("GetChainRoa(): roaId:", roaId, "  chainRoa.Id:", chainRoa.Id)

	// get current stateModel
	chainRoa.StateModel = model.GetStateModelAndResetStage(chainRoa.State, "chainvalidate")
	belogs.Debug("GetChainRoa():stateModel:", chainRoa.StateModel)

	//lab_rpki_roa_ipaddress
	belogs.Debug("GetChainRoa():certChainCert.id:", chainRoa.Id)
	chainRoa.ChainIpAddresses, chainRoa.ChainEeIpAddresses, err = getChainRoaIpAddresses(chainRoa.Id)
	if err != nil {
		belogs.Error("GetChainRoa(): getChainIpAddresses fail, chainRoa.Id:", chainRoa.Id, err)
		return chainRoa, err
	}

	// get sn in crl
	chainRoa.ChainSnInCrlRevoked, err = getRoaEeSnInCrlRevoked(roaId)
	if err != nil {
		belogs.Error("GetChainRoa(): getRoaEeSnInCrlRevoked fail, chainRoa.Id:", chainRoa.Id, err)
		return chainRoa, err
	}

	belogs.Debug("GetChainRoa(): roaId:", roaId, "   chainRoa.Id:", chainRoa.Id, "  time(s):", time.Now().Sub(start).Seconds())
	return chainRoa, nil
}

func getChainRoaIpAddresses(roaId uint64) (chainIpAddresses, chainEeIpAddress []chainmodel.ChainIpAddress, err error) {
	start := time.Now()

	err = xormdb.XormEngine.Table("lab_rpki_roa_ipaddress").
		Cols("id,addressFamily,addressPrefix,maxLength,rangeStart,rangeEnd").
		Where("roaId=?", roaId).
		OrderBy("id").Find(&chainIpAddresses)
	if err != nil {
		belogs.Error("getChainIpAddresses(): lab_rpki_roa_ipaddress fail:", err)
		return nil, nil, err
	}
	err = xormdb.XormEngine.Table("lab_rpki_roa_ee_ipaddress").
		Cols("id,addressFamily,addressPrefix,min,max,rangeStart,rangeEnd,addressPrefixRange").
		Where("roaId=?", roaId).
		OrderBy("id").Find(&chainEeIpAddress)
	if err != nil {
		belogs.Error("getChainIpAddresses(): lab_rpki_roa_ipaddress fail:", err)
		return nil, nil, err
	}

	belogs.Debug("getChainIpAddresses():roaId, len(chainIpAddresses),len(chainEeIpAddress):",
		roaId, len(chainIpAddresses), len(chainEeIpAddress))
	belogs.Debug("getChainIpAddresses(): roaId:", roaId, "  time(s):", time.Now().Sub(start).Seconds())
	return chainIpAddresses, chainEeIpAddress, nil
}

func getRoaEeSnInCrlRevoked(roaId uint64) (chainSnInCrlRevoked chainmodel.ChainSnInCrlRevoked, err error) {
	start := time.Now()
	sql := `select l.fileName, r.revocationTime from lab_rpki_roa c, lab_rpki_crl l, lab_rpki_crl_revoked_cert r
	 where  c.jsonAll->'$.eeCertModel.sn' = r.sn and r.crlId = l.id and c.aki = l.aki and c.id=` + convert.ToString(roaId)
	belogs.Debug("getRoaEeSnInCrlRevoked(): roaId:", roaId, "   sql:", sql)
	_, err = xormdb.XormEngine.
		Sql(sql).Get(&chainSnInCrlRevoked)
	if err != nil {
		belogs.Error("getRoaEeSnInCrlRevoked(): select fail:", roaId, err)
		return chainSnInCrlRevoked, err
	}
	belogs.Debug("getRoaEeSnInCrlRevoked(): roaId:", roaId, "  time(s):", time.Now().Sub(start).Seconds())
	return chainSnInCrlRevoked, nil

}

func UpdateRoas(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	roaIds := chains.RoaIds
	for _, roaId := range roaIds {

		err = updateRoa(session, chains, roaId)
		if err != nil {
			belogs.Error("UpdateRoas(): updateRoa fail, roaId:", roaId, err)
			xormdb.RollbackAndLogError(session, "UpdateRoas(): updateRoa fail: "+convert.ToString(roaId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRoas(): CommitSession fail :", err)
		return
	}
	belogs.Debug("UpdateRoas(): len(roaIds):", len(roaIds), "  time(s):", time.Now().Sub(start).Seconds())

}

func updateRoa(session *xorm.Session, chains *chainmodel.Chains, roaId uint64) (err error) {
	start := time.Now()
	chainRoa, err := chains.GetRoaById(roaId)
	if err != nil {
		belogs.Error("updateRoa(): GetRoa fail :", roaId, err)
		return err
	}

	chainDbRoaModel := chainmodel.NewChainDbRoaModel(&chainRoa)
	originModel := model.JudgeOrigin(chainRoa.FilePath)

	chainCerts := jsonutil.MarshalJson(*chainDbRoaModel)
	state := jsonutil.MarshalJson(chainRoa.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateRoa():roaId:", roaId, "    chainCerts:", chainCerts, "    origin:", origin, "  state:", state)
	sqlStr := `UPDATE lab_rpki_roa set chainCerts=?, state=?, origin=?   where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, roaId)
	if err != nil {
		belogs.Error("updateRoa(): UPDATE lab_rpki_roa fail :", roaId, err)
		return err
	}
	belogs.Debug("updateRoa():roaId:", roaId, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
