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

func GetChainRoaSqls() (chainCertSqls []chainmodel.ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]chainmodel.ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_roa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("GetChainRoaSqls(): lab_rpki_roa id fail:", err)
		return nil, err
	}
	belogs.Info("GetChainRoaSqls(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Now().Sub(start).Seconds())
	return chainCertSqls, nil
}

func UpdateRoas(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
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
