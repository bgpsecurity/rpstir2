package sync

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

// state: chainValidating;
func updateSyncLogRtrStateStartDb(state string) (labRpkiSyncLogId uint64, err error) {
	start := time.Now()
	belogs.Debug("updateSyncLogRtrStateStartDb():  state:", state)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateSyncLogRtrStateStartDb(): NewSession fail :", err)
		return 0, err
	}
	defer session.Close()

	var id int64
	_, err = session.SQL("SELECT MAX(id) AS id FROM lab_rpki_sync_log").Get(&id) // session.Table("lab_rpki_sync_log").Select(`"max(id)" as id`).Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateSyncLogRtrStateStartDb(): update sync_log fail: state:"+state, err)
	}
	syncLogRtrState := model.SyncLogRtrState{
		StartTime: time.Now(),
	}
	rtrState := jsonutil.MarshalJson(syncLogRtrState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set rtrState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, rtrState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateSyncLogRtrStateStartDb(): UPDATE sync_log fail: rtrState:"+
			rtrState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateSyncLogRtrStateStartDb(): CommitSession fail:"+
			rtrState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	belogs.Info("updateSyncLogRtrStateStartDb(): CommitSession ok:   state:", state, "   time(s):", time.Since(start))
	return uint64(id), nil
}

func updateSyncLogRtrStateEndDb(labRpkiSyncLogId uint64, state string) (err error) {
	// get current rtrState, the set new value
	start := time.Now()
	belogs.Debug("updateSyncLogRtrStateEndDb(): labRpkiSyncLogId:", labRpkiSyncLogId, " state:", state)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateSyncLogRtrStateEndDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	var rtrState string
	_, err = session.Table("lab_rpki_sync_log").Cols("rtrState").Where("id = ?", labRpkiSyncLogId).Get(&rtrState)
	if err != nil {
		belogs.Error("updateSyncLogRtrStateEndDb(): lab_rpki_sync_log Get rtrState :", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "updateSyncLogRtrStateEndDb(): CommitSession fail: ", err)
	}

	syncLogRtrState := model.SyncLogRtrState{}
	jsonutil.UnmarshalJson(rtrState, &syncLogRtrState)
	syncLogRtrState.EndTime = time.Now()
	rtrState = jsonutil.MarshalJson(syncLogRtrState)
	belogs.Debug("updateSyncLogRtrStateEndDb():syncLogRtrState:", rtrState)

	sqlStr := `UPDATE lab_rpki_sync_log set rtrState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, rtrState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("updateSyncLogRtrStateEndDb(): UPDATE sync_log fail : rtrState: ",
			rtrState, "   state:", state, "    labRpkiSyncLogId:", (labRpkiSyncLogId), err)
		return xormdb.RollbackAndLogError(session, "updateSyncLogRtrStateEndDb(): CommitSession fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateSyncLogRtrStateEndDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateSyncLogRtrStateEndDb(): CommitSession fail: ", err)
	}
	belogs.Info("updateSyncLogRtrStateEndDb(): CommitSession ok: labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state,
		"   time(s):", time.Since(start))
	return nil
}
