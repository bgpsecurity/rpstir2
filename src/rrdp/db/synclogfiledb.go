package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
)

func InsertSyncLogFile(session *xorm.Session,
	syncLogId uint64,
	syncType, filePath, fileName, fileHash string, rrdpTime time.Time) error {
	belogs.Debug("InsertSyncLogFile():rrdp,  syncLogId, rrdpTime, filePath, fileName,fileHash, rrdpTime:",
		syncLogId, filePath, fileName, fileHash, rrdpTime)

	fileType := osutil.ExtNoDot(fileName)
	rtr := "notNeed"
	if osutil.ExtNoDot(fileName) == "roa" {
		rtr = "notYet"
	}

	labRpkiSyncLogFileState := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "notYet",
		Rtr:             rtr,
	}
	state := jsonutil.MarshalJson(labRpkiSyncLogFileState)
	belogs.Debug("InsertSyncLogFile():rrdp,fileType,rrdpTime,filePath,fileName,syncType,state:",
		fileType, rrdpTime, filePath, fileName, syncType, state)

	//lab_rpki_sync_log_file
	// when delta ,may change same file ,in more than on delta files,
	// so insert ignore into, just save one time
	sqlStr := `INSERT ignore into lab_rpki_sync_log_file
	               (syncLogId, fileType,syncTime,
	               filePath,fileName,syncType,
	               syncStyle,state,fileHash)
			 VALUES(?,?,?,
			 ?,?,?,
			 ?,?,?)`
	_, err := session.Exec(sqlStr, syncLogId, fileType, rrdpTime,
		filePath, fileName, syncType,
		"rrdp", state, xormdb.SqlNullString(fileHash))
	if err != nil {
		belogs.Error("InsertSyncLogFile():rrdp, INSERT lab_rpki_sync_log_file fail, syncLogId:", syncLogId, filePath, fileName, err)
		return err
	}
	return nil
}
