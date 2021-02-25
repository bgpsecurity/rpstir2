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

func InsertSyncLogFiles(labRpkiSyncLogId uint64,
	addFiles, delFiles, updateFiles map[string]rsyncmodel.RsyncFileHash) (err error) {
	belogs.Debug("InsertSyncLogFiles():rsync, labRpkiSyncLogId:", labRpkiSyncLogId)

	// update lab_rpki_sync_log
	session, err := xormdb.NewSession()
	defer session.Close()

	rsyncTime := time.Now()

	// insert lab_rpki_sync_log_file
	rsyncType := "add"
	for _, fileHash := range addFiles {
		err = InsertSyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): InsertRsyncLogFile add fail : fileName:"+
				fileHash.FileName, err)
		}
	}
	belogs.Debug("InsertSyncLogFiles():rsync, after len(addFiles):", labRpkiSyncLogId, len(addFiles), "   addFiles:", jsonutil.MarshalJson(addFiles))

	rsyncType = "update"
	for _, fileHash := range updateFiles {
		err = InsertSyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): InsertRsyncLogFile update fail : fileName:"+
				fileHash.FileName, err)
		}
	}
	belogs.Debug("InsertSyncLogFiles():rsync, after len(updateFiles):", labRpkiSyncLogId, len(updateFiles), "   updateFiles:", jsonutil.MarshalJson(updateFiles))

	rsyncType = "del"
	for _, fileHash := range delFiles {
		err = InsertSyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "InsertRsyncLogFiles(): InsertRsyncLogFile  del fail :  fileName:"+
				fileHash.FileName, err)
		}
	}
	belogs.Debug("InsertSyncLogFiles():rsync, after len(delFiles):", labRpkiSyncLogId, len(delFiles), "   delFiles:", jsonutil.MarshalJson(delFiles))

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "InsertSyncLogFiles(): CommitSession fail: labRpkiSyncLogId:"+
			convert.ToString(labRpkiSyncLogId), err)
	}
	return nil

}

func InsertSyncLogFile(session *xorm.Session, labRpkiSyncLogId uint64, rsyncTime time.Time,
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
	//lab_rpki_sync_log_file
	sqlStr := `INSERT ignore lab_rpki_sync_log_file
	               (syncLogId, fileType,syncTime,
	               filePath,fileName,syncType,
	               syncStyle,state,fileHash)
			 VALUES(?,?,?,
			 ?,?,?,
			 ?,?,?)`
	_, err := session.Exec(sqlStr, labRpkiSyncLogId, rsyncFileHash.FileType, rsyncTime,
		rsyncFileHash.FilePath, rsyncFileHash.FileName, syncType,
		"rsync", state, xormdb.SqlNullString(rsyncFileHash.FileHash))
	belogs.Debug("InsertSyncLogFile(): rsync, labRpkiSyncLogId:",
		labRpkiSyncLogId, "    rsyncFileHash:", rsyncFileHash.FilePath, rsyncFileHash.FileName,
		"   syncType:", syncType, "  state:", state)
	if err != nil {
		belogs.Error("InsertSyncLogFile(): rsync, INSERT lab_rpki_sync_log_file fail, labRpkiSyncLogId:", labRpkiSyncLogId, rsyncFileHash.FilePath, rsyncFileHash.FileName, err)
		return xormdb.RollbackAndLogError(session, "INSERT lab_rpki_sync_log_file fail", err)
	}
	return nil
}
