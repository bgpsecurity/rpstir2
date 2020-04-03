package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
	rsyncmodel "rsync/model"
)

// state: rsyncing;  style: rsync
func InsertRsyncLogRsyncStateStart(state, syncStyle string) (logId uint64, err error) {
	syncLogRsyncState := model.SyncLogRsyncState{
		StartTime: time.Now(),
	}
	rsyncState := jsonutil.MarshalJson(syncLogRsyncState)

	session, err := xormdb.NewSession()
	defer session.Close()
	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_log(rsyncState, state,syncStyle)
					VALUES(?,?,?)`
	res, err := session.Exec(sqlStr, rsyncState, state, syncStyle)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "InsertRsyncLogRsyncStateStart(): INSERT lab_rpki_sync_log fail:"+rsyncState+","+state+","+syncStyle, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "InsertRsyncLogRsyncStateStart(): LastInsertId fail:"+rsyncState+","+state+","+syncStyle, err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "InsertRsyncLogRsyncStateStart(): CommitSession fail:"+rsyncState+","+state+","+syncStyle, err)

	}

	belogs.Debug("InsertRsyncLogRsyncStateStart():LastInsertId:", id)
	return uint64(id), nil
}

func UpdateRsyncLogRsyncStateEnd(labRpkiSyncLogId uint64, state string, misc *rsyncmodel.RsyncMisc) (err error) {
	belogs.Debug("UpdateRsyncLogRsyncStateEnd():labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state, "   misc:", misc)

	session, err := xormdb.NewSession()
	defer session.Close()

	// get current rsyncState, the set new value
	var rsyncState string
	_, err = session.Table("lab_rpki_sync_log").Cols("rsyncState").Where("id = ?", labRpkiSyncLogId).Get(&rsyncState)
	if err != nil {
		belogs.Error("UpdateRsyncLogRsyncStateEnd(): lab_rpki_sync_log Get rsyncState :", labRpkiSyncLogId, err)
		return err
	}
	syncLogRsyncState := model.SyncLogRsyncState{}
	jsonutil.UnmarshalJson(rsyncState, &syncLogRsyncState)
	syncLogRsyncState.EndTime = time.Now()
	syncLogRsyncState.FailRsyncUrls = misc.FailRsyncUrls
	syncLogRsyncState.OkRsyncUrlLen = misc.OkRsyncUrlLen
	rsyncState = jsonutil.MarshalJson(syncLogRsyncState)
	belogs.Debug("UpdateRsyncLogRsyncStateEnd():syncLogRsyncState:", rsyncState)

	sqlStr := `UPDATE lab_rpki_sync_log set rsyncState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, rsyncState, state, labRpkiSyncLogId)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRsyncStateEnd(): UPDATE lab_rpki_sync_log fail : rsyncState: "+
			rsyncState+"   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRsyncStateEnd(): CommitSession fail:"+
			rsyncState+","+state+","+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	return nil
}

// state: rsyncing;  style: rsync
func UpdateRsyncLogDiffStateStart(labRpkiSyncLogId uint64, state string) (err error) {
	belogs.Debug("UpdateRsyncLogDiffStateStart():labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state)

	syncLogDiffState := model.SyncLogDiffState{
		StartTime: time.Now(),
	}
	diffState := jsonutil.MarshalJson(syncLogDiffState)

	session, err := xormdb.NewSession()
	defer session.Close()
	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set diffState=?, state=?  where id=?`
	_, err = session.Exec(sqlStr, diffState, state, labRpkiSyncLogId)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogDiffStateStart(): INSERT lab_rpki_sync_log fail: diffState:"+
			diffState+",   state:"+state+",    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogDiffStateStart(): CommitSession fail:"+
			diffState+","+state+",    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	return nil
}

func UpdateRsyncLogDiffStateEnd(labRpkiSyncLogId uint64, state string, filesFromDb,
	filesFromDisk, addFiles, delFiles, updateFiles, noChangeFiles map[string]rsyncmodel.RsyncFileHash) (err error) {
	belogs.Debug("UpdateRsyncLogDiffStateEnd():labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state)

	// update lab_rpki_sync_log
	session, err := xormdb.NewSession()
	defer session.Close()
	var diffState string
	_, err = session.Table("lab_rpki_sync_log").Cols("diffState").Where("id = ?", labRpkiSyncLogId).Get(&diffState)
	if err != nil {
		belogs.Error("UpdateRsyncLogRsyncStateEnd(): lab_rpki_sync_log Get diffState :", labRpkiSyncLogId, err)
		return err
	}
	rsyncTime := time.Now()

	syncLogDiffState := model.SyncLogDiffState{}
	jsonutil.UnmarshalJson(diffState, &syncLogDiffState)
	syncLogDiffState.AddFilesLen = uint64(len(addFiles))
	syncLogDiffState.DelFilesLen = uint64(len(delFiles))
	syncLogDiffState.EndTime = rsyncTime
	syncLogDiffState.FilesFromDbLen = uint64(len(filesFromDb))
	syncLogDiffState.FilesFromDiskLen = uint64(len(filesFromDisk))
	syncLogDiffState.NoChangeFilesLen = uint64(len(noChangeFiles))
	syncLogDiffState.UpdateFilesLen = uint64(len(updateFiles))
	diffState = jsonutil.MarshalJson(syncLogDiffState)
	belogs.Debug("UpdateRsyncLogRsyncStateEnd():diffState:", diffState)

	sqlStr := `UPDATE lab_rpki_sync_log set diffState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, diffState, state, labRpkiSyncLogId)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRsyncStateEnd(): UPDATE lab_rpki_sync_log fail : diffState: "+
			diffState+"   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	// insert lab_rpki_sync_log_file

	rsyncType := "add"
	for _, fileHash := range addFiles {
		err = InsertRsyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRsyncStateEnd(): UPDATE lab_rpki_sync_log add fail : diffState: "+
				diffState+"   state:"+state+"    fileName:"+fileHash.FileName, err)
		}
	}
	rsyncType = "update"
	for _, fileHash := range updateFiles {
		err = InsertRsyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRsyncStateEnd(): UPDATE lab_rpki_sync_log update fail : diffState: "+
				diffState+"   state:"+state+"    fileName:"+fileHash.FileName, err)
		}
	}
	rsyncType = "del"
	for _, fileHash := range delFiles {
		err = InsertRsyncLogFile(session, labRpkiSyncLogId, rsyncTime, rsyncType, &fileHash)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRsyncStateEnd(): UPDATE lab_rpki_sync_log del fail : diffState: "+
				diffState+"   state:"+state+"    fileName:"+fileHash.FileName, err)
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogDiffStateStart(): CommitSession fail:"+
			diffState+","+state+",    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	return nil

}
