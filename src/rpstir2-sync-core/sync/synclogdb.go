package sync

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

// syncStyle: sync/rsync/rrdp,state: syncing;
func InsertSyncLogStartDb(syncStyle string, state string) (syncLogId uint64, err error) {

	syncLogSyncState := model.SyncLogSyncState{StartTime: time.Now(), SyncStyle: syncStyle}
	syncState := jsonutil.MarshalJson(syncLogSyncState)
	belogs.Debug("InsertSyncLogStartDb():syncStyle:", syncStyle, "   state:", state, "  syncState:", syncState)

	session, err := xormdb.NewSession()
	defer session.Close()
	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_log(syncState, state,syncStyle)
					VALUES(?,?,?)`
	res, err := session.Exec(sqlStr, syncState, state, syncStyle)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSyncLogStartDb(): INSERT lab_rpki_sync_log fail:"+syncState+","+state+","+syncStyle, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSyncLogStartDb(): LastInsertId fail:"+syncState+","+state+","+syncStyle, err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSyncLogStartDb(): CommitSession fail:"+syncState+","+state+","+syncStyle, err)

	}

	belogs.Debug("InsertSyncLogStartDb():new syncLogId:", id)
	return uint64(id), nil
}

// state: synced
func UpdateSyncLogEndDb(labRpkiSyncLogId uint64, state string, syncState string) (err error) {
	start := time.Now()
	belogs.Debug("UpdateSyncLogEndDb():labRpkiSyncLogId:", labRpkiSyncLogId,
		"   state:", state, "   syncState:", syncState)

	session, err := xormdb.NewSession()
	defer session.Close()

	sqlStr := `UPDATE lab_rpki_sync_log set state=?, syncState=?  where id=? `
	_, err = session.Exec(sqlStr, state, syncState, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("UpdateSyncLogEndDb(): UPDATE lab_rpki_sync_log fail : syncState: "+
			syncState+"   state:"+state, "    labRpkiSyncLogId:", labRpkiSyncLogId, err,
			"time(s):", time.Since(start))
		return xormdb.RollbackAndLogError(session, "UpdateSyncLogEndDb(): UPDATE lab_rpki_sync_log fail", err)
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateSyncLogEndDb(): CommitSession fail : syncState: "+
			syncState+"   state:"+state, "    labRpkiSyncLogId:", labRpkiSyncLogId, err,
			"time(s):", time.Since(start))
		return xormdb.RollbackAndLogError(session, "UpdateSyncLogEndDb(): CommitSession fail:", err)
	}
	belogs.Info("UpdateSyncLogEndDb(): ok, labRpkiSyncLogId:", labRpkiSyncLogId,
		"   state:", state, "time(s):", time.Since(start))
	return nil
}
