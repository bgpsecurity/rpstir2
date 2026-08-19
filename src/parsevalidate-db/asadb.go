package parsevalidatedb

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func InsertAsaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("InsertAsaDb(): will add asa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertAsaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertAsaDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("InsertAsaDb(): InsertAsaDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertAsaDb(): InsertAsaDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("InsertAsaDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertAsaDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("InsertAsaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("InsertAsaDb(): asa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func InsertAsaDbWithSession(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	if syncLogFileModel.CertModel == nil {
		belogs.Error("InsertAsaDbWithSession(): CertModel is nil, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is nil")
	}

	var asaModel model.AsaModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	belogs.Debug("InsertAsaDbWithSession():asaModel json:", json)
	err := jsonutil.UnmarshalJson(json, &asaModel)
	if err != nil {
		belogs.Error("InsertAsaDbWithSession(): json fail, CertModel to asaModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not asaModel type")
	}

	orginModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	belogs.Debug("InsertAsaDbWithSession():asaModel filePath,fileName:", asaModel.FilePath, asaModel.FileName, "  orginModel:", orginModelJson, "  now ", now)

	//lab_rpki_asa
	sqlStr := `INSERT INTO lab_rpki_asa(
	                ski, aki, filePath, fileName, 
	                fileHash, jsonAll, syncLogId, syncLogFileId, updateTime,
	                state, origin)
					VALUES(?,?,?,?,
					?,?,?,?,?,
					?,?)`
	res, err := session.Exec(sqlStr,
		xormdb.SqlNullString(asaModel.Ski), xormdb.SqlNullString(asaModel.Aki), asaModel.FilePath, asaModel.FileName,
		asaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(asaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(orginModelJson))
	if err != nil {
		belogs.Error("InsertAsaDbWithSession(): INSERT INTO lab_rpki_asa Exec fail,",
			"  asaModel:", asaModel.String(), " syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	asaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertAsaDbWithSession(): LastInsertId fail:", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_asa_aia
	belogs.Debug("InsertAsaDbWithSession(): asaId:", asaId, " asaModel.Aia.CaIssuers:", asaModel.AiaModel.CaIssuers)
	if len(asaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT INTO lab_rpki_asa_aia(asaId, caIssuers)
				VALUES(?,?)`
		_, err = session.Exec(sqlStr, asaId, asaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertAsaDbWithSession(): INSERT INTO lab_rpki_asa_aia fail, asaId:", asaId, "  CaIssuers:", asaModel.AiaModel.CaIssuers, err)
			return err
		}
	}

	//lab_rpki_asa_sia
	belogs.Debug("InsertAsaDbWithSession(): asaModel.Sia:", asaModel.SiaModel)
	if len(asaModel.SiaModel.CaRepository) > 0 ||
		len(asaModel.SiaModel.RpkiManifest) > 0 ||
		len(asaModel.SiaModel.RpkiNotify) > 0 ||
		len(asaModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT INTO lab_rpki_asa_sia(asaId, rpkiManifest,rpkiNotify,caRepository,signedObject)
				VALUES(?,?,?,?,?)`
		_, err = session.Exec(sqlStr, asaId, asaModel.SiaModel.RpkiManifest,
			asaModel.SiaModel.RpkiNotify, asaModel.SiaModel.CaRepository,
			asaModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertAsaDbWithSession(): INSERT INTO lab_rpki_asa_sia fail, SiaModel:", jsonutil.MarshalJson(asaModel.SiaModel), err)
			return err
		}
	}

	//lab_rpki_asa_customer_provider_asn
	belogs.Debug("InsertAsaDbWithSession(): asaModel:", asaModel.String())
	customerAndProviderSqlStr := `INSERT INTO lab_rpki_asa_customer_provider_asn(
				asaId,customerAsn, providerAsn, providerAsnOrder) 
				VALUES(?,?,?,?)`
	if asaModel.CustomerAsns != nil && len(asaModel.CustomerAsns) > 0 {
		for _, customerAsn := range asaModel.CustomerAsns {
			cAsn := customerAsn.CustomerAsn

			for i, pAsn := range customerAsn.ProviderAsns {
				_, err = session.Exec(customerAndProviderSqlStr,
					asaId, cAsn, pAsn, i)
				if err != nil {
					belogs.Error("InsertAsaDbWithSession(): INSERT INTO lab_rpki_asa_customer_provider_asn fail:",
						"  asaId:", asaId, "  customerAsn:", cAsn, "  providerAsn:", pAsn, "  providerAsnOrder:", i, err)
					return err
				}
			}
		}
	}
	belogs.Debug("InsertAsaDbWithSession(): insert asaModel ok,filePath,fileName:", asaModel.FilePath, asaModel.FileName)
	return nil
}

func DelAsaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelAsaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	err = DelAsaDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("DelAsaDb(): DelAsaDbWithSession fail:", err)
		return xormdb.RollbackAndLogError(session, "DelAsaDb(): DelAsaDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelAsaDb(): CommitSession fail :", err)
		return err
	}
	return nil
}

func DelAsaDbWithSession(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelAsaDbWithSession(): will del asa file, certId:", syncLogFileModel.CertId,
		" filePath:", syncLogFileModel.FilePath, "  fileName:", syncLogFileModel.FileName)

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathNameWithSession(session, "lab_rpki_asa", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelAsaDbWithSession(): getCertIdByFilePathNameWithSession fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return err
		}
		if certId == 0 {
			belogs.Info("DelAsaDbWithSession(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelAsaDbWithSession(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelAsaByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelAsaDbWithSession(): DelAsaByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in InsertAsaDb()
	if syncLogFileModel.SyncType == "del" {
		err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelAsaDbWithSession(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
	}

	belogs.Info("DelAsaDbWithSession(): asa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelAsaByIdDbWithSession(session *xorm.Session, asaId uint64) (err error) {

	belogs.Debug("DelAsaByIdDbWithSession():delete lab_rpki_asa by asaId:", asaId)

	// rrdp may have id==0, just return nil
	if asaId <= 0 {
		return nil
	}
	belogs.Info("DelAsaByIdDbWithSession():delete lab_rpki_asa by asaId, more than 0:", asaId)

	//lab_rpki_asa_customer_provider_asn
	res, err := session.Exec("delete from lab_rpki_asa_customer_provider_asn  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDbWithSession():delete  from lab_rpki_asa_customer_provider_asn fail: asaId: ", asaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelAsaByIdDbWithSession():delete lab_rpki_asa_customer_provider_asn by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_aia
	res, err = session.Exec("delete from  lab_rpki_asa_aia  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDbWithSession():delete  from lab_rpki_asa_aia fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDbWithSession():delete lab_rpki_asa_aia by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_sia
	res, err = session.Exec("delete from  lab_rpki_asa_sia  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDbWithSession():delete  from lab_rpki_asa_sia fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDbWithSession():delete lab_rpki_asa_sia by asaId:", asaId, "  count:", count)

	//lab_rpki_asa
	res, err = session.Exec("delete from  lab_rpki_asa  where id = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDbWithSession():delete  from lab_rpki_asa fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDbWithSession():delete lab_rpki_asa by asaId:", asaId, "  count:", count)

	return nil
}

/*
func getExpireAsaDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
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

func updateAsaStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateAsaStateDb(): NewSession fail:", err)
		return err
	}
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
	belogs.Info("updateAsaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateAsaByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireAsaDb(now)
	if err != nil {
		belogs.Error("UpdateAsaByCheckAll(): getExpireAsaDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateAsaByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateAsaByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NotAfter of EE is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter of EE is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleEe") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateAsaByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateAsaStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateAsaByCheckAll(): updateAsaStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateAsaByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
*/
