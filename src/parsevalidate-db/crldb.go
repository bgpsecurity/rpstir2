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

func InsertCrlDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("InsertCrlDb(): will add crl file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertCrlDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertCrlDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("InsertCrlDb(): InsertCrlDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertCrlDb(): InsertCrlDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("InsertCrlDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertCrlDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("InsertCrlDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("InsertCrlDb(): crl file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func InsertCrlDbWithSession(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {
	belogs.Debug("InsertCrlDbWithSession(): file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	if syncLogFileModel.CertModel == nil {
		belogs.Error("InsertCrlDbWithSession(): CertModel is nil, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is nil")
	}

	var crlModel model.CrlModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	belogs.Debug("InsertCrlDbWithSession():crlModel json:", json)
	err := jsonutil.UnmarshalJson(json, &crlModel)
	if err != nil {
		belogs.Error("InsertCrlDbWithSession(): json fail, CertModel to crlModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not crlModel type")
	}
	originModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	thisUpdate := crlModel.ThisUpdate
	nextUpdate := crlModel.NextUpdate
	belogs.Debug("InsertCrlDbWithSession(): crlModel filePath,fileName:", crlModel.FilePath, crlModel.FileName,
		"  origin:", originModelJson, "  now ", now)

	//lab_rpki_crl
	sqlStr := `INSERT INTO lab_rpki_crl(
	        crlNumber, thisUpdate, nextUpdate, hasExpired, aki, 
	        filePath, fileName, fileHash, jsonAll, syncLogId, 
	        syncLogFileId, updateTime, state, origin)
			VALUES(?,?,?,?,?,
			?,?,?,?,?,
			?,?,?,?)`
	res, err := session.Exec(sqlStr,
		crlModel.CrlNumber, thisUpdate, nextUpdate, crlModel.HasExpired, xormdb.SqlNullString(crlModel.Aki),
		crlModel.FilePath, crlModel.FileName, crlModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(crlModel)), syncLogFileModel.SyncLogId,
		syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(originModelJson))
	if err != nil {
		belogs.Error("InsertCrlDbWithSession(): INSERT INTO lab_rpki_crl Exec fail:",
			"  crlModel:", crlModel.String(), "  syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	crlId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertCrlDbWithSession(): LastInsertId fail:", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_crl_crlrevokedcerts
	belogs.Debug("InsertCrlDbWithSession(): crlId:", crlId, " len(crlModel.RevokedCertModels):", len(crlModel.RevokedCertModels))
	if crlModel.RevokedCertModels != nil && len(crlModel.RevokedCertModels) > 0 {
		sqlStr = `INSERT INTO lab_rpki_crl_revoked_cert(crlId, sn, revocationTime) VALUES(?,?,?)`
		for _, revokedCertModel := range crlModel.RevokedCertModels {
			_, err = session.Exec(sqlStr, crlId, revokedCertModel.Sn, revokedCertModel.RevocationTime)
			if err != nil {
				belogs.Error("InsertCrlDbWithSession(): INSERT INTO lab_rpki_crl_revoked_cert Exec :",
					syncLogFileModel.String(), err)
				return err
			}
		}
	}
	return nil
}

func DelCrlDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelCrlDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	err = DelCrlDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("DelCrlDb(): DelCrlDbWithSession fail:", err)
		return xormdb.RollbackAndLogError(session, "DelCrlDb(): DelCrlDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelCrlDb(): CommitSession fail :", err)
		return err
	}
	return nil
}

func DelCrlDbWithSession(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelCrlDbWithSession(): will del crl file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathNameWithSession(session, "lab_rpki_crl", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelCrlDbWithSession(): getCertIdByFilePathNameWithSession fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return err
		}
		if certId == 0 {
			belogs.Info("DelCrlDbWithSession(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelCrlDbWithSession(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelCrlByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelCrlDbWithSession(): DelCrlByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in InsertCrlDbWithSession()
	if syncLogFileModel.SyncType == "del" {
		err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelCrlDbWithSession(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
	}

	belogs.Info("DelCrlDbWithSession(): crl file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelCrlByIdDbWithSession(session *xorm.Session, crlId uint64) (err error) {
	belogs.Debug("DelCrlByIdDbWithSession():delete lab_rpki_crl by crlId:", crlId)

	// rrdp may have id==0, just return nil
	if crlId <= 0 {
		return nil
	}
	belogs.Info("DelCrlByIdDbWithSession():delete lab_rpki_crl by crlId, more than 0:", crlId)

	//lab_rpki_crl_revoked_cert
	res, err := session.Exec("delete from lab_rpki_crl_revoked_cert  where crlId = ?", crlId)
	if err != nil {
		belogs.Error("DelCrlByIdDbWithSession():delete  from lab_rpki_crl_revoked_cert fail: crlId: ", crlId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelCrlByIdDbWithSession():delete lab_rpki_crl_revoked_cert by crlId:", crlId, "  count:", count)

	//lab_rpki_crl
	res, err = session.Exec("delete from  lab_rpki_crl  where id = ?", crlId)
	if err != nil {
		belogs.Error("DelCrlByIdDbWithSession():delete  from lab_rpki_crl fail: crlId: ", crlId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCrlByIdDbWithSession():delete lab_rpki_crl by crlId:", crlId, "  count:", count)

	return nil

}

/*
func getExpireCrlDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.nextUpdate  as endTime  from  lab_rpki_crl c
			where timestamp(c.nextUpdate) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireCrlDb(): lab_rpki_crl fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireCrlDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateCrlStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateCrlStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_crl c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateCrlStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateCrlStateDb(): UPDATE lab_rpki_crl fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateCrlStateDb(): UPDATE lab_rpki_crl fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateCrlStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateCrlStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateCrlByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireCrlDb(now)
	if err != nil {
		belogs.Error("UpdateCrlByCheckAll(): getExpireCrlDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateCrlByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateCrlByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NextUpdate is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleCrl") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateCrlByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateCrlStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateCrlByCheckAll(): updateCrlStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateCrlByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
*/
