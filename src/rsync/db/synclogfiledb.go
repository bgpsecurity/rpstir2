package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
	rsyncmodel "rsync/model"
)

func InsertRsyncLogFile(session *xorm.Session, labRpkiSyncLogId uint64, rsyncTime time.Time,
	syncType string, rsyncFileHash *rsyncmodel.RsyncFileHash) error {

	rtr := "notNeed"
	if rsyncFileHash.FileType == "roa" {
		rtr = "notYet"
	}

	labRpkiSyncLogFileState := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "notYet",
		Rtr:             rtr,
	}
	state := jsonutil.MarshalJson(labRpkiSyncLogFileState)
	belogs.Debug("InsertRsyncLogFile():  labRpkiSyncLogId:", labRpkiSyncLogId, "  state:", state)

	//lab_rpki_sync_log_file
	sqlStr := `INSERT lab_rpki_sync_log_file
	               (syncLogId, fileType,syncTime,
	               filePath,fileName,syncType,
	               state,fileHash)
			 VALUES(?,?,?,
			 ?,?,?,
			 ?,?)`
	_, err := session.Exec(sqlStr, labRpkiSyncLogId, rsyncFileHash.FileType, rsyncTime,
		rsyncFileHash.FilePath, rsyncFileHash.FileName, syncType,
		state, xormdb.SqlNullString(rsyncFileHash.FileHash))
	if err != nil {
		belogs.Error("InsertRsyncLogFile(): INSERT lab_rpki_sync_log_file fail, labRpkiSyncLogId:", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "INSERT lab_rpki_sync_log_file fail", err)
	}
	return nil
}
