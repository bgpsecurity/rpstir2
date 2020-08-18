package db

import (
	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

// state: syncing;  style: rsync/rrdp
func InsertSyncLogSyncStateStart(state, syncStyle string, syncLogSyncState *model.SyncLogSyncState) (logId uint64, err error) {

	syncState := jsonutil.MarshalJson(syncLogSyncState)

	session, err := xormdb.NewSession()
	defer session.Close()
	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_log(syncState, state,syncStyle)
					VALUES(?,?,?)`
	res, err := session.Exec(sqlStr, syncState, state, syncStyle)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSyncLogSyncStateStart(): INSERT lab_rpki_sync_log fail:"+syncState+","+state+","+syncStyle, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSyncLogSyncStateStart(): LastInsertId fail:"+syncState+","+state+","+syncStyle, err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSyncLogSyncStateStart(): CommitSession fail:"+syncState+","+state+","+syncStyle, err)

	}

	belogs.Debug("InsertSyncLogSyncStateStart():LastInsertId:", id)
	return uint64(id), nil
}

// state: synced
func UpdateSyncLogSyncStateEnd(labRpkiSyncLogId uint64, state string, syncLogSyncState *model.SyncLogSyncState) (err error) {
	belogs.Debug("UpdateSyncLogSyncStateEnd():labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	syncState := jsonutil.MarshalJson(syncLogSyncState)
	belogs.Debug("UpdateSyncLogSyncStateEnd():syncState:", syncState)

	sqlStr := `UPDATE lab_rpki_sync_log set syncState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, syncState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("UpdateSyncLogSyncStateEnd(): UPDATE lab_rpki_sync_log fail : syncState: "+
			syncState+"   state:"+state, "    labRpkiSyncLogId:", convert.ToString(labRpkiSyncLogId), err)
		return err
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateSyncLogSyncStateEnd(): CommitSession fail:", err)
	}
	return nil
}
