package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
	rsyncmodel "rsync/model"
)

func InsertRsyncLogFiles(labRpkiSyncLogId uint64,
	addFiles, delFiles, updateFiles map[string]rsyncmodel.RsyncFileHash) (err error) {
	belogs.Debug("InsertRsyncLogFiles():labRpkiSyncLogId:", labRpkiSyncLogId)

	// update lab_rpki_sync_log
	session, err := xormdb.NewSession()
	defer session.Close()

	rsyncTime := time.Now()

	// insert lab_rpki_sync_log_file
	rsyncType := "add"
	for _, fileHash := range addFiles {
		err = InsertRsyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): InsertRsyncLogFile add fail : fileName:"+
				fileHash.FileName, err)
		}
	}
	rsyncType = "update"
	for _, fileHash := range updateFiles {
		err = InsertRsyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): InsertRsyncLogFile update fail : fileName:"+
				fileHash.FileName, err)
		}
	}
	rsyncType = "del"
	for _, fileHash := range delFiles {
		err = InsertRsyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): InsertRsyncLogFile  del fail :  fileName:"+
				fileHash.FileName, err)
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): CommitSession fail: labRpkiSyncLogId:"+
			convert.ToString(labRpkiSyncLogId), err)
	}
	return nil

}

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
	               syncStyle,state,fileHash)
			 VALUES(?,?,?,
			 ?,?,?,
			 ?,?,?)`
	_, err := session.Exec(sqlStr, labRpkiSyncLogId, rsyncFileHash.FileType, rsyncTime,
		rsyncFileHash.FilePath, rsyncFileHash.FileName, syncType,
		"rsync", state, xormdb.SqlNullString(rsyncFileHash.FileHash))
	if err != nil {
		belogs.Error("InsertRsyncLogFile(): INSERT lab_rpki_sync_log_file fail, labRpkiSyncLogId:", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "INSERT lab_rpki_sync_log_file fail", err)
	}
	return nil
}
