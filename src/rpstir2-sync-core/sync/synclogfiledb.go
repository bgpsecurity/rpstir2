package sync

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
)

func InsertSyncLogFilesDb(syncLogFiles []model.SyncLogFile) error {
	belogs.Debug("InsertSyncLogFilesDb(): len(syncLogFiles):", len(syncLogFiles))
	if len(syncLogFiles) == 0 {
		return nil
	}

	session, err := xormdb.NewSession()
	defer session.Close()

	//lab_rpki_sync_log_file
	sqlStr := `INSERT ignore lab_rpki_sync_log_file
	               (syncLogId,fileType,syncTime,
	               filePath,fileName,sourceUrl,
				   syncType,syncStyle,state,
				   fileHash)
			 VALUES(?,?,?,
			 ?,?,?,
			 ?,?,?,
			 ?)`
	for i := range syncLogFiles {
		_, err := session.Exec(sqlStr,
			syncLogFiles[i].SyncLogId, syncLogFiles[i].FileType, syncLogFiles[i].SyncTime,
			syncLogFiles[i].FilePath, syncLogFiles[i].FileName, syncLogFiles[i].SourceUrl,
			syncLogFiles[i].SyncType, syncLogFiles[i].SyncStyle, syncLogFiles[i].State,
			xormdb.SqlNullString(syncLogFiles[i].FileHash))
		belogs.Debug("InsertSyncLogFilesDb(): syncLogFiles:", jsonutil.MarshalJson(syncLogFiles[i]))
		if err != nil {
			belogs.Error("InsertSyncLogFilesDb(): rsync, INSERT lab_rpki_sync_log_file fail:", jsonutil.MarshalJson(syncLogFiles[i]), err)
			return xormdb.RollbackAndLogError(session, "INSERT lab_rpki_sync_log_file fail", err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "InsertSyncLogFilesDb(): CommitSession fail:", err)
	}
	return nil
}
