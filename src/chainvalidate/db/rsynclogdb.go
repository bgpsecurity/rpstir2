package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

// state: chainValidating;
func UpdateRsyncLogChainValidateStateStart(state string) (labRpkiSyncLogId uint64, err error) {
	belogs.Debug("UpdateRsyncLogChainValidateStateStart():  state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	var id int64
	_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogChainValidateStateStart(): update lab_rpki_sync_log fail: state:"+state, err)
	}
	syncLogChainValidateState := model.SyncLogChainValidateState{
		StartTime: time.Now(),
	}
	chainValidateState := jsonutil.MarshalJson(syncLogChainValidateState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set chainValidateState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, chainValidateState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogChainValidateStateStart(): UPDATE lab_rpki_sync_log fail: chainValidateState:"+
			chainValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogChainValidateStateStart(): CommitSession fail:"+
			chainValidateState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	return uint64(id), nil
}

func UpdateRsyncLogChainValidateStateEnd(labRpkiSyncLogId uint64, state string) (err error) {
	// get current chainValidateState, the set new value
	session, err := xormdb.NewSession()
	defer session.Close()

	var chainValidateState string
	_, err = session.Table("lab_rpki_sync_log").Cols("chainValidateState").Where("id = ?", labRpkiSyncLogId).Get(&chainValidateState)
	if err != nil {
		belogs.Error("UpdateRsyncLogChainValidateStateEnd(): lab_rpki_sync_log Get parseValidateState :", labRpkiSyncLogId, err)
		return err
	}
	syncLogChainValidateState := model.SyncLogChainValidateState{}
	jsonutil.UnmarshalJson(chainValidateState, &syncLogChainValidateState)
	syncLogChainValidateState.EndTime = time.Now()
	chainValidateState = jsonutil.MarshalJson(syncLogChainValidateState)
	belogs.Debug("UpdateRsyncLogChainValidateStateEnd():syncLogChainValidateState:", syncLogChainValidateState)

	sqlStr := `UPDATE lab_rpki_sync_log set chainValidateState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, chainValidateState, state, labRpkiSyncLogId)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogChainValidateStateEnd(): UPDATE lab_rpki_sync_log fail : chainValidateState: "+
			chainValidateState+"   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogChainValidateStateEnd(): CommitSession fail:"+
			chainValidateState+","+state+","+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	return nil
}
