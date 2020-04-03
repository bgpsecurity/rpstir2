package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

func GetSyncLogFileModels(labRpkiSyncLogId uint64, syncType, fileType string) (syncLogFileModels []model.SyncLogFileModel, err error) {
	// get lastest syncLogFile.Id

	belogs.Debug("GetSyncLogFileModels():start")
	syncLogFileModels = make([]model.SyncLogFileModel, 0)
	err = xormdb.XormEngine.Table("lab_rpki_sync_log_file").Select("id,syncLogId,filePath,fileName, fileType, syncType").
		Where("state->'$.updateCertTable'=?", "notYet").And("syncLogId=?", labRpkiSyncLogId).
		And("syncType=?", syncType).
		And("fileType=?", fileType).
		OrderBy("id").Find(&syncLogFileModels)
	if err != nil {
		belogs.Error("GetSyncLogFileModels(): Find fail:", err)
		return nil, err
	}
	belogs.Debug("GetSyncLogFileModels(): len(syncLogFileModels):", len(syncLogFileModels))

	var certId uint64
	var tableName string
	for i, _ := range syncLogFileModels {
		switch syncLogFileModels[i].FileType {
		case "cer":
			tableName = "lab_rpki_cer"
		case "crl":
			tableName = "lab_rpki_crl"
		case "mft":
			tableName = "lab_rpki_mft"
		case "roa":
			tableName = "lab_rpki_roa"
		}
		has, err := xormdb.XormEngine.Table(tableName).Where("filePath=?", syncLogFileModels[i].FilePath).
			And("fileName=?", syncLogFileModels[i].FileName).Cols("id").Get(&certId)
		if err != nil {
			belogs.Error("GetSyncLogFileModels(): get id fail:", tableName,
				syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName, err)
			return nil, err
		}
		if has {
			syncLogFileModels[i].CertId = certId
		}
		belogs.Debug("GetSyncLogFileModels():get id: ", tableName,
			syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName, syncLogFileModels[i].CertId)
	}
	return syncLogFileModels, nil
}

// state: parseValidating;
func UpdateRsyncLogParseValidateStart(state string) (labRpkiSyncLogId uint64, err error) {
	belogs.Debug("UpdateRsyncLogParseValidateStart():  state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	var id int64
	_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogParseValidateStart(): update lab_rpki_sync_log fail: state:"+state, err)
	}
	syncLogParseValidateState := model.SyncLogParseValidateState{
		StartTime: time.Now(),
	}
	parseValidateState := jsonutil.MarshalJson(syncLogParseValidateState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set parseValidateState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, parseValidateState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogParseValidateStart(): UPDATE lab_rpki_sync_log fail: parseValidateState:"+
			parseValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogParseValidateStart(): CommitSession fail:"+
			parseValidateState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	return uint64(id), nil
}
func UpdateRsyncLogParseValidateStateEnd(labRpkiSyncLogId uint64, state string,
	parseFailFiles []string) (err error) {
	session, err := xormdb.NewSession()
	defer session.Close()

	// get current rsyncState, the set new value
	var parseValidateState string
	_, err = session.Table("lab_rpki_sync_log").Cols("parseValidateState").Where("id = ?", labRpkiSyncLogId).Get(&parseValidateState)
	if err != nil {
		belogs.Error("UpdateRsyncLogParseValidateStateEnd(): lab_rpki_sync_log Get parseValidateState :", labRpkiSyncLogId, err)
		return err
	}
	syncLogParseValidateState := model.SyncLogParseValidateState{}
	jsonutil.UnmarshalJson(parseValidateState, &syncLogParseValidateState)
	syncLogParseValidateState.EndTime = time.Now()
	syncLogParseValidateState.ParseFailFiles = parseFailFiles
	parseValidateState = jsonutil.MarshalJson(syncLogParseValidateState)
	belogs.Debug("UpdateRsyncLogParseValidateStateEnd():parseValidateState:", parseValidateState)

	sqlStr := `UPDATE lab_rpki_sync_log set parseValidateState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, parseValidateState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("UpdateRsyncLogParseValidateStateEnd(): lab_rpki_sync_log UPDATE :", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogParseValidateStateEnd(): UPDATE lab_rpki_sync_log fail: parseValidateState:"+
			parseValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRsyncLogParseValidateStateEnd(): CommitSession fail:"+
			parseValidateState+","+state+",  labRpkiSyncLogId:", labRpkiSyncLogId, err)
		return err
	}
	return nil
}
