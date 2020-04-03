package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	"github.com/go-xorm/xorm"

	"model"
)

func InsertRsyncLogFile(session *xorm.Session,
	labRpkiSyncLogId uint64,
	url, syncType, pathFileName string, rrdpTime time.Time) error {
	belogs.Debug("InsertRsyncLogFile():  labRpkiSyncLogId, rrdpTime, pathFileName:",
		labRpkiSyncLogId, rrdpTime, pathFileName)

	filePath, fileName := osutil.GetFilePathAndFileName(pathFileName)
	fileType := osutil.ExtNoDot(fileName)

	rtr := "notNeed"
	if osutil.ExtNoDot(url) == "roa" {
		rtr = "notYet"
	}

	labRpkiSyncLogFileState := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "notYet",
		Rtr:             rtr,
	}
	state := jsonutil.MarshalJson(labRpkiSyncLogFileState)
	belogs.Debug("InsertRsyncLogFile():fileType,rrdpTime,filePath,fileName,syncType,state:",
		fileType, rrdpTime, filePath, fileName, syncType, state)

	//lab_rpki_sync_log_file
	// when delta ,may change same file ,in more than on delta files,
	// so insert ignore into, just save one time
	sqlStr := `INSERT ignore into lab_rpki_sync_log_file
	               (syncLogId, fileType,syncTime,
	               filePath,fileName,syncType,
	               state)
			 VALUES(?,?,?,
			 ?,?,?,
			 ?)`
	_, err := session.Exec(sqlStr, labRpkiSyncLogId, fileType, rrdpTime,
		filePath, fileName, syncType,
		state)
	if err != nil {
		belogs.Error("InsertRsyncLogFile(): INSERT lab_rpki_sync_log_file fail, labRpkiSyncLogId:", labRpkiSyncLogId, err)
		return err
	}
	return nil
}
