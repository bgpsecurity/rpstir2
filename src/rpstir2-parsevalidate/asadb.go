package parsevalidate

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

// add
func addAsasDb(syncLogFileModels []SyncLogFileModel) error {
	belogs.Info("addAsasDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addAsasDb(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new asa
	for i := range syncLogFileModels {
		err = insertAsaDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addAsasDb(): insertAsaDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addAsasDb(): insertAsaDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail ", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addAsasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addAsasDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

// del
func delAsasDb(delSyncLogFileModels []SyncLogFileModel, updateSyncLogFileModels []SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delAsasDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delAsaByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delAsasDb(): delAsaByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delAsasDb(): delAsaByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delAsasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delAsasDb(): len(asas), ", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func delAsaByIdDb(session *xorm.Session, asaId uint64) (err error) {

	belogs.Debug("delAsaByIdDb():delete lab_rpki_asa by asaId:", asaId)

	// rrdp may have id==0, just return nil
	if asaId <= 0 {
		return nil
	}
	belogs.Info("delAsaByIdDb():delete lab_rpki_asa by asaId, more than 0:", asaId)

	//lab_rpki_asa_provider_asn
	res, err := session.Exec("delete from lab_rpki_asa_provider_asn  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("delAsaByIdDb():delete  from lab_rpki_asa_provider_asn fail: asaId: ", asaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("delAsaByIdDb():delete lab_rpki_asa_provider_asn by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_customer_asn
	res, err = session.Exec("delete from lab_rpki_asa_customer_asn  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("delAsaByIdDb():delete  from lab_rpki_asa_customer_asn fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delAsaByIdDb():delete lab_rpki_asa_customer_asn by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_aia
	res, err = session.Exec("delete from  lab_rpki_asa_aia  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("delAsaByIdDb():delete  from lab_rpki_asa_aia fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delAsaByIdDb():delete lab_rpki_asa_aia by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_sia
	res, err = session.Exec("delete from  lab_rpki_asa_sia  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("delAsaByIdDb():delete  from lab_rpki_asa_sia fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delAsaByIdDb():delete lab_rpki_asa_sia by asaId:", asaId, "  count:", count)

	//lab_rpki_asa
	res, err = session.Exec("delete from  lab_rpki_asa  where id = ?", asaId)
	if err != nil {
		belogs.Error("delAsaByIdDb():delete  from lab_rpki_asa fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delAsaByIdDb():delete lab_rpki_asa by asaId:", asaId, "  count:", count)

	return nil
}

func insertAsaDb(session *xorm.Session,
	syncLogFileModel *SyncLogFileModel, now time.Time) error {

	asaModel := syncLogFileModel.CertModel.(model.AsaModel)
	//lab_rpki_asa
	sqlStr := `INSERT lab_rpki_asa(
	                ski, aki, filePath,fileName, 
	                fileHash,jsonAll,syncLogId, syncLogFileId, updateTime,
	                state)
					VALUES(?,?,?,?,
					?,?,?,?,?,
					?)`
	res, err := session.Exec(sqlStr,
		xormdb.SqlNullString(asaModel.Ski), xormdb.SqlNullString(asaModel.Aki), asaModel.FilePath, asaModel.FileName,
		asaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(asaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)))
	if err != nil {
		belogs.Error("insertAsaDb(): INSERT lab_rpki_asa Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	asaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertAsaDb(): LastInsertId asaId:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_asa_aia
	belogs.Debug("insertAsaDb(): asaModel.Aia.CaIssuers:", asaModel.AiaModel.CaIssuers)
	if len(asaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_asa_aia(asaId, caIssuers)
				VALUES(?,?)`
		res, err = session.Exec(sqlStr, asaId, asaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertAsaDb(): INSERT lab_rpki_asa_aia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_asa_sia
	belogs.Debug("insertAsaDb(): asaModel.Sia:", asaModel.SiaModel)
	if len(asaModel.SiaModel.CaRepository) > 0 ||
		len(asaModel.SiaModel.RpkiManifest) > 0 ||
		len(asaModel.SiaModel.RpkiNotify) > 0 ||
		len(asaModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_asa_sia(asaId, rpkiManifest,rpkiNotify,caRepository,signedObject)
				VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, asaId, asaModel.SiaModel.RpkiManifest,
			asaModel.SiaModel.RpkiNotify, asaModel.SiaModel.CaRepository,
			asaModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("insertAsaDb(): INSERT lab_rpki_asa_sia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_asa_customer_asn
	belogs.Debug("insertAsaDb(): asaModel.CustomerAsns:", jsonutil.MarshalJson(asaModel.CustomerAsns))
	if asaModel.CustomerAsns != nil && len(asaModel.CustomerAsns) > 0 {
		customerSqlStr := `INSERT lab_rpki_asa_customer_asn(asaId, addressFamily,customerAsn)
						VALUES(?,?,?)`
		for _, customerAsn := range asaModel.CustomerAsns {
			res, err = session.Exec(customerSqlStr, asaId, customerAsn.AddressFamily, customerAsn.CustomerAsn)
			if err != nil {
				belogs.Error("insertAsaDb(): INSERT lab_rpki_asa_customer_asn Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
			customerAsnId, err := res.LastInsertId()
			if err != nil {
				belogs.Error("insertAsaDb(): LastInsertId customerAsnId:", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}

			//lab_rpki_asa_provider_asn
			belogs.Debug("insertAsaDb(): customerAsn.ProviderAsns:", jsonutil.MarshalJson(customerAsn.ProviderAsns))
			providerSqlStr := `INSERT lab_rpki_asa_provider_asn(asaId,customerAsnId, providerAsn,addressFamily,providerOrder) 
	                 VALUES(?,?,?,?,?)`
			for i, providerAsn := range customerAsn.ProviderAsns {
				res, err = session.Exec(providerSqlStr,
					asaId, customerAsnId, providerAsn.ProviderAsn, providerAsn.AddressFamily, i)
				if err != nil {
					belogs.Error("insertAsaDb(): INSERT lab_rpki_asa_provider_asn Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
					return err
				}
			}
		}
	}
	return nil
}

func getExpireAsaDb(now time.Time) (certIdStateModels []CertIdStateModel, err error) {

	certIdStateModels = make([]CertIdStateModel, 0)
	t := now.Local().Format("2006-01-02T15:04:05-0700")
	sql := `select id, state as stateStr,str_to_date( SUBSTRING_INDEX(c.jsonAll->>'$.eeCertModel.notAfter','+',1),'%Y-%m-%dT%H:%i:%S')  as endTime  from  lab_rpki_asa c 
			where c.jsonAll->>'$.eeCertModel.notAfter' < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireAsaDb(): lab_rpki_asa fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireAsaDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateAsaStateDb(certIdStateModels []CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update lab_rpki_asa c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateAsaStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateAsaStateDb(): UPDATE lab_rpki_asa fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateAsaStateDb(): UPDATE lab_rpki_asa fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateAsaStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateAsaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Now().Sub(start))

	return nil
}
