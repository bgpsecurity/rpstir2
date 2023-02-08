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

func getChainAsaSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_asa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainAsaSqlsDb(): lab_rpki_asa id fail:", err)
		return nil, err
	}
	belogs.Info("getChainAsaSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func updateAsasDb(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	asaIds := chains.AsaIds
	for _, asaId := range asaIds {

		err = updateAsaDb(session, chains, asaId)
		if err != nil {
			belogs.Error("updateAsasDb(): updateAsaDb fail, asaId:", asaId, err)
			xormdb.RollbackAndLogError(session, "updateAsasDb(): updateAsaDb fail: "+convert.ToString(asaId), err)
			return
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateAsasDb(): CommitSession fail :", err)
		return
	}
	belogs.Debug("updateAsasDb(): len(asaIds):", len(asaIds), "  time(s):", time.Since(start))

}

func updateAsaDb(session *xorm.Session, chains *Chains, asaId uint64) (err error) {
	start := time.Now()
	chainAsa, err := chains.GetAsaById(asaId)
	if err != nil {
		belogs.Error("updateAsaDb(): GetAsa fail :", asaId, err)
		return err
	}

	chainDbAsaModel := NewChainDbAsaModel(&chainAsa)
	originModel := model.JudgeOrigin(chainAsa.FilePath)

	chainCerts := jsonutil.MarshalJson(*chainDbAsaModel)
	state := jsonutil.MarshalJson(chainAsa.StateModel)
	origin := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateAsaDb():asaId:", asaId, "    chainCerts:", chainCerts, "    origin:", origin, "  state:", state)
	sqlStr := `UPDATE lab_rpki_asa set chainCerts=?, state=?, origin=?   where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, origin, asaId)
	if err != nil {
		belogs.Error("updateAsaDb(): UPDATE lab_rpki_asa fail :", asaId, err)
		return err
	}
	belogs.Debug("updateAsaDb():asaId:", asaId, "  time(s):", time.Since(start))
	return nil
}
