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

func InsertMftDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("InsertMftDb(): will add mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertMftDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertMftDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("InsertMftDb(): InsertMftDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertMftDb(): InsertMftDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("InsertMftDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertMftDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("InsertMftDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("InsertMftDb(): mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func InsertMftDbWithSession(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	if syncLogFileModel.CertModel == nil {
		belogs.Error("InsertMftDbWithSession(): CertModel is nil, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is nil")
	}

	var mftModel model.MftModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	belogs.Debug("InsertMftDbWithSession():mftModel json:", json)
	err := jsonutil.UnmarshalJson(json, &mftModel)
	if err != nil {
		belogs.Error("InsertMftDbWithSession(): json fail, CertModel to mftModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not mftModel type")
	}
	originModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	thisUpdate := mftModel.ThisUpdate
	nextUpdate := mftModel.NextUpdate
	belogs.Debug("InsertMftDbWithSession(): mftModel filePath,fileName:", mftModel.FilePath, mftModel.FileName, " originModel:", originModelJson, " now ", now, "  thisUpdate:", thisUpdate, "  nextUpdate:", nextUpdate)

	//lab_rpki_manifest
	sqlStr := `INSERT INTO lab_rpki_mft(
	           mftNumber, thisUpdate, nextUpdate, ski, aki, 
	           filePath, fileName, fileHash, jsonAll, syncLogId, 
	           syncLogFileId, updateTime, state, origin)
				VALUES(
				?,?,?,?,?,
				?,?,?,?,?,
				?,?,?,?)`
	res, err := session.Exec(sqlStr,
		mftModel.MftNumber, thisUpdate, nextUpdate, xormdb.SqlNullString(mftModel.Ski), xormdb.SqlNullString(mftModel.Aki),
		mftModel.FilePath, mftModel.FileName, mftModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(mftModel)), syncLogFileModel.SyncLogId,
		syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(originModelJson))
	if err != nil {
		belogs.Error("InsertMftDbWithSession(): INSERT INTO lab_rpki_mft Exec fail, ",
			"  mftModel:", mftModel.String(), "  syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	mftId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertMftDbWithSession(): LastInsertId fail:", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_mft_aia
	belogs.Debug("InsertMftDbWithSession(): mftId:", mftId, "  mftModel.Aia.CaIssuers:", mftModel.AiaModel.CaIssuers)
	if len(mftModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT INTO lab_rpki_mft_aia(mftId, caIssuers) 
			VALUES(?,?)`
		_, err = session.Exec(sqlStr, mftId, mftModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertMftDbWithSession(): INSERT INTO lab_rpki_mft_aia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_mft_sia
	belogs.Debug("InsertMftDbWithSession(): mftModel.Sia:", mftModel.SiaModel)
	if len(mftModel.SiaModel.CaRepository) > 0 ||
		len(mftModel.SiaModel.RpkiManifest) > 0 ||
		len(mftModel.SiaModel.RpkiNotify) > 0 ||
		len(mftModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT INTO lab_rpki_mft_sia(mftId, rpkiManifest,rpkiNotify,caRepository,signedObject) 
			VALUES(?,?,?,?,?)`
		_, err = session.Exec(sqlStr, mftId, mftModel.SiaModel.RpkiManifest,
			mftModel.SiaModel.RpkiNotify, mftModel.SiaModel.CaRepository,
			mftModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertMftDbWithSession(): INSERT INTO lab_rpki_mft_sia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_mft_fileAndHashs
	belogs.Debug("InsertMftDbWithSession(): len(mftModel.FileHashModels):", len(mftModel.FileHashModels))
	if mftModel.FileHashModels != nil && len(mftModel.FileHashModels) > 0 {
		sqlStr = `INSERT INTO lab_rpki_mft_file_hash(mftId, file,hash) VALUES(?,?,?)`
		for _, fileHashModel := range mftModel.FileHashModels {
			_, err = session.Exec(sqlStr, mftId, fileHashModel.File, fileHashModel.Hash)
			if err != nil {
				belogs.Error("InsertMftDbWithSession(): INSERT INTO lab_rpki_mft_file_hash Exec :", syncLogFileModel.String(), err)
				return err
			}
		}
	}
	return nil
}

func DelMftDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelMftDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	err = DelMftDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("DelMftDb(): DelMftDbWithSession fail:", err)
		return xormdb.RollbackAndLogError(session, "DelMftDb(): DelMftDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelMftDb(): CommitSession fail :", err)
		return err
	}
	return nil
}

func DelMftDbWithSession(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelMftDbWithSession(): will del mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathNameWithSession(session, "lab_rpki_mft", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelMftDbWithSession(): getCertIdByFilePathNameWithSession fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return err
		}
		if certId == 0 {
			belogs.Info("DelMftDbWithSession(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelMftDbWithSession(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelMftByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelMftDbWithSession(): DelMftByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in InsertMftDbWithSession()
	if syncLogFileModel.SyncType == "del" {
		err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelMftDbWithSession(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
	}

	belogs.Info("DelMftDbWithSession(): mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelMftByIdDbWithSession(session *xorm.Session, mftId uint64) (err error) {
	belogs.Debug("DelMftByIdDbWithSession():delete lab_rpki_mft by mftId:", mftId)

	// rrdp may have id==0, just return nil
	if mftId <= 0 {
		return nil
	}
	belogs.Info("DelMftByIdDbWithSession():delete lab_rpki_mft by mftId, more than 0:", mftId)

	//lab_rpki_mft_file_hash
	res, err := session.Exec("delete from lab_rpki_mft_file_hash  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDbWithSession():delete  from lab_rpki_mft_file_hash fail: mftId: ", mftId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelMftByIdDbWithSession():delete lab_rpki_mft_file_hash by mftId:", mftId, "  count:", count)

	//lab_rpki_mft_sia
	res, err = session.Exec("delete from  lab_rpki_mft_sia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDbWithSession():delete  from lab_rpki_mft_sia fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelMftByIdDbWithSession():delete lab_rpki_mft_sia by mftId:", mftId, "  count:", count)

	//lab_rpki_mft_aia
	res, err = session.Exec("delete from  lab_rpki_mft_aia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDbWithSession():delete  from lab_rpki_mft_aia fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelMftByIdDbWithSession():delete lab_rpki_mft_aia by mftId:", mftId, "  count:", count)

	//lab_rpki_mft
	res, err = session.Exec("delete from  lab_rpki_mft  where id = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDbWithSession():delete  from lab_rpki_mft fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelMftByIdDbWithSession():delete lab_rpki_mft by mftId:", mftId, "  count:", count)

	return nil

}

/*
func getExpireMftDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.nextUpdate  as endTime from  lab_rpki_mft c
			where timestamp(c.nextUpdate) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireMftDb(): lab_rpki_mft fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireMftDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateMftStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateMftStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_mft c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateMftStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateMftStateDb(): UPDATE lab_rpki_mft fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateMftStateDb(): UPDATE lab_rpki_mft fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateMftStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateMftStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateMftByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireMftDb(now)
	if err != nil {
		belogs.Error("UpdateMftByCheckAll(): getExpireMftDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateMftByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateMftByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NextUpdate is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleMft") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateMftByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateMftStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateMftByCheckAll(): updateMftStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateMftByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
*/
