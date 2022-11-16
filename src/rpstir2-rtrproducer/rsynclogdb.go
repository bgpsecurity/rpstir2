package rtrproducer

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
)

// state: chainValidating;
func updateRsyncLogRtrStateStartDb(state string) (labRpkiSyncLogId uint64, err error) {
	start := time.Now()
	belogs.Debug("updateRsyncLogRtrStateStartDb():  state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	var id int64
	_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateRsyncLogRtrStateStartDb(): update lab_rpki_sync_log fail: state:"+state, err)
	}
	syncLogRtrState := model.SyncLogRtrState{
		StartTime: time.Now(),
	}
	rtrState := jsonutil.MarshalJson(syncLogRtrState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set rtrState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, rtrState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateRsyncLogRtrStateStartDb(): UPDATE lab_rpki_sync_log fail: rtrState:"+
			rtrState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateRsyncLogRtrStateStartDb(): CommitSession fail:"+
			rtrState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	belogs.Info("updateRsyncLogRtrStateStartDb(): CommitSession ok:   state:", state, "   time(s):", time.Since(start))
	return uint64(id), nil
}

func updateRsyncLogRtrStateEndDb(labRpkiSyncLogId uint64, state string) (err error) {
	// get current rtrState, the set new value
	start := time.Now()
	belogs.Debug("updateRsyncLogRtrStateEndDb(): labRpkiSyncLogId:", labRpkiSyncLogId, " state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	var rtrState string
	_, err = session.Table("lab_rpki_sync_log").Cols("rtrState").Where("id = ?", labRpkiSyncLogId).Get(&rtrState)
	if err != nil {
		belogs.Error("updateRsyncLogRtrStateEndDb(): lab_rpki_sync_log Get rtrState :", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "updateRsyncLogRtrStateEndDb(): CommitSession fail: ", err)
	}

	syncLogRtrState := model.SyncLogRtrState{}
	jsonutil.UnmarshalJson(rtrState, &syncLogRtrState)
	syncLogRtrState.EndTime = time.Now()
	rtrState = jsonutil.MarshalJson(syncLogRtrState)
	belogs.Debug("updateRsyncLogRtrStateEndDb():syncLogRtrState:", rtrState)

	sqlStr := `UPDATE lab_rpki_sync_log set rtrState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, rtrState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("updateRsyncLogRtrStateEndDb(): UPDATE lab_rpki_sync_log fail : rtrState: ",
			rtrState, "   state:", state, "    labRpkiSyncLogId:", (labRpkiSyncLogId), err)
		return xormdb.RollbackAndLogError(session, "updateRsyncLogRtrStateEndDb(): CommitSession fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRsyncLogRtrStateEndDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateRsyncLogRtrStateEndDb(): CommitSession fail: ", err)
	}
	belogs.Info("updateRsyncLogRtrStateEndDb(): CommitSession ok: labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state,
		"   time(s):", time.Since(start))
	return nil
}
