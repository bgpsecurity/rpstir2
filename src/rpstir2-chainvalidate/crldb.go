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

func GetChainCrlSqls() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.jsonAll, c.state ,cer.cerFiles,roa.roaFiles, mft.mftFiles 
		from lab_rpki_crl c  
		left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as cerFiles , v.id as crlId from lab_rpki_cer c, lab_rpki_crl_revoked_cert_view v 
			 where c.sn = v.sn and c.aki =v.aki 
			 group by v.id) cer on cer.crlId = c.id	
		left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as roaFiles , v.id as crlId from lab_rpki_roa c, lab_rpki_crl_revoked_cert_view v 
			 where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
			 group by v.id) roa on roa.crlId = c.id 
		left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as mftFiles , v.id as crlId from lab_rpki_mft c, lab_rpki_crl_revoked_cert_view v 
			 where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
			 group by v.id) mft on mft.crlId = c.id	 `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("GetChainCrlSqls(): lab_rpki_crl id fail:", err)
		return nil, err
	}
	belogs.Info("GetChainCrlSqls(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Now().Sub(start).Seconds())
	return chainCertSqls, nil
}

func UpdateCrls(chains *Chains, wg *sync.WaitGroup) {
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

func updateCrl(session *xorm.Session, chains *Chains, crlId uint64) (err error) {

	start := time.Now()
	chainCrl, err := chains.GetCrlById(crlId)
	if err != nil {
		belogs.Error("updateCrl(): GetCrl fail :", crlId, err)
		return err
	}

	chainDbCrlModel := NewChainDbCrlModel(&chainCrl)
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
