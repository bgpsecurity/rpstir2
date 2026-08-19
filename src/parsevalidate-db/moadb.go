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

func InsertMoaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("InsertMoaDb(): will add moa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertMoaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertMoaDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("InsertMoaDb(): InsertMoaDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertMoaDb(): InsertMoaDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("InsertMoaDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertMoaDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("InsertMoaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("InsertMoaDb(): moa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func InsertMoaDbWithSession(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	if syncLogFileModel.CertModel == nil {
		belogs.Error("InsertMoaDbWithSession(): CertModel is nil, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is nil")
	}

	var moaModel model.MoaModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	belogs.Debug("InsertMoaDbWithSession():moaModel json:", json)
	err := jsonutil.UnmarshalJson(json, &moaModel)
	if err != nil {
		belogs.Error("InsertMoaDbWithSession(): json fail, CertModel to moaModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not moaModel type")
	}

	orginModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	ipv4PrefixesStr := ""
	if len(moaModel.Ipv4Prefixes) > 0 {
		ipv4PrefixesStr = jsonutil.MarshalJson(moaModel.Ipv4Prefixes)
	}
	belogs.Debug("InsertMoaDbWithSession():moaModel filePath,fileName:",
		moaModel.FilePath, moaModel.FileName, "  orginModel:", orginModelJson,
		"ipv4PrefixesStr", ipv4PrefixesStr, " now ", now)

	//lab_rpki_moa
	sqlStr := `INSERT INTO lab_rpki_moa(
	                ski, aki, filePath, fileName, 
	                fileHash, jsonAll, syncLogId, syncLogFileId, updateTime,
					ipv6MappingPrefix, ipv4Prefixes,
	                state, origin)
					VALUES(?,?,?,?,
					?,?,?,?,?,
					?,?,
					?,?)`
	res, err := session.Exec(sqlStr,
		xormdb.SqlNullString(moaModel.Ski), xormdb.SqlNullString(moaModel.Aki), moaModel.FilePath, moaModel.FileName,
		moaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(moaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		moaModel.Ipv6MappingPrefix, xormdb.SqlNullString(ipv4PrefixesStr),
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)), xormdb.SqlNullString(orginModelJson))
	if err != nil {
		belogs.Error("InsertMoaDbWithSession(): INSERT INTO lab_rpki_moa Exec fail,",
			"  moaModel:", moaModel.String(), " syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	moaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertMoaDbWithSession(): LastInsertId fail:", syncLogFileModel.String(), err)
		return err
	}

	belogs.Debug("InsertMoaDbWithSession(): insert moaModel ok,filePath,fileName:", moaModel.FilePath, moaModel.FileName, "  moaId:", moaId)
	return nil
}

func DelMoaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelMoaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	err = DelMoaDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("DelMoaDb(): DelMoaDbWithSession fail:", err)
		return xormdb.RollbackAndLogError(session, "DelMoaDb(): DelMoaDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelMoaDb(): CommitSession fail :", err)
		return err
	}
	return nil
}

func DelMoaDbWithSession(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelMoaDbWithSession(): will del moa file, certId:", syncLogFileModel.CertId,
		" filePath:", syncLogFileModel.FilePath, "  fileName:", syncLogFileModel.FileName)

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathNameWithSession(session, "lab_rpki_moa", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelMoaDbWithSession(): getCertIdByFilePathNameWithSession fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return err
		}
		if certId == 0 {
			belogs.Info("DelMoaDbWithSession(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelMoaDbWithSession(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelMoaByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelMoaDbWithSession(): DelMoaByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in InsertMoaDb()
	if syncLogFileModel.SyncType == "del" {
		err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelMoaDbWithSession(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
	}

	belogs.Info("DelMoaDbWithSession(): moa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelMoaByIdDbWithSession(session *xorm.Session, moaId uint64) (err error) {

	belogs.Debug("DelMoaByIdDbWithSession():delete lab_rpki_moa by moaId:", moaId)

	// rrdp may have id==0, just return nil
	if moaId <= 0 {
		return nil
	}
	belogs.Info("DelMoaByIdDbWithSession():delete lab_rpki_moa by moaId, more than 0:", moaId)

	//lab_rpki_moa
	res, err := session.Exec("delete from  lab_rpki_moa  where id = ?", moaId)
	if err != nil {
		belogs.Error("DelMoaByIdDbWithSession():delete  from lab_rpki_moa fail: moaId: ", moaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelMoaByIdDbWithSession():delete lab_rpki_moa by moaId:", moaId, "  count:", count)

	return nil
}

/*
func getExpireMoaDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := now.Local().Format("2006-01-02T15:04:05-0700")
	sql := `select id, state as stateStr,str_to_date( SUBSTRING_INDEX(c.jsonAll->>'$.eeCertModel.notAfter','+',1),'%Y-%m-%dT%H:%i:%S')  as endTime  from  lab_rpki_moa c
			where c.jsonAll->>'$.eeCertModel.notAfter' < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireMoaDb(): lab_rpki_moa fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireMoaDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateMoaStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateMoaStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_moa c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateMoaStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateMoaStateDb(): UPDATE lab_rpki_moa fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateMoaStateDb(): UPDATE lab_rpki_moa fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateMoaStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateMoaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateMoaByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireMoaDb(now)
	if err != nil {
		belogs.Error("UpdateMoaByCheckAll(): getExpireMoaDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateMoaByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateMoaByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
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
		belogs.Debug("UpdateMoaByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateMoaStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateMoaByCheckAll(): updateMoaStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateMoaByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
*/
