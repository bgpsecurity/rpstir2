package parsevalidate

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
)

// state: parseValidating;
func updateRsyncLogParseValidateStartDb(state string) (labRpkiSyncLogId uint64, err error) {
	belogs.Debug("updateRsyncLogParseValidateStartDb():  state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	var id int64
	_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateRsyncLogParseValidateStartDb(): update lab_rpki_sync_log fail: state:"+state, err)
	}
	syncLogParseValidateState := model.SyncLogParseValidateState{
		StartTime: time.Now(),
	}
	parseValidateState := jsonutil.MarshalJson(syncLogParseValidateState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set parseValidateState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, parseValidateState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateRsyncLogParseValidateStartDb(): UPDATE lab_rpki_sync_log fail: parseValidateState:"+
			parseValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "updateRsyncLogParseValidateStartDb(): CommitSession fail:"+
			parseValidateState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	return uint64(id), nil
}
func updateRsyncLogParseValidateStateEndDb(labRpkiSyncLogId uint64, state string,
	parseFailFiles []string) (err error) {
	session, err := xormdb.NewSession()
	defer session.Close()

	// get current parseValidateState, the set new value
	var parseValidateState string
	_, err = session.Table("lab_rpki_sync_log").Cols("parseValidateState").Where("id = ?", labRpkiSyncLogId).Get(&parseValidateState)
	if err != nil {
		belogs.Error("updateRsyncLogParseValidateStateEndDb(): lab_rpki_sync_log Get parseValidateState :", labRpkiSyncLogId, err)
		return err
	}
	syncLogParseValidateState := model.SyncLogParseValidateState{}
	jsonutil.UnmarshalJson(parseValidateState, &syncLogParseValidateState)
	syncLogParseValidateState.EndTime = time.Now()
	syncLogParseValidateState.ParseFailFiles = parseFailFiles
	parseValidateState = jsonutil.MarshalJson(syncLogParseValidateState)
	belogs.Debug("updateRsyncLogParseValidateStateEndDb():parseValidateState:", parseValidateState)

	sqlStr := `UPDATE lab_rpki_sync_log set parseValidateState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, parseValidateState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("updateRsyncLogParseValidateStateEndDb(): lab_rpki_sync_log UPDATE :", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "updateRsyncLogParseValidateStateEndDb(): UPDATE lab_rpki_sync_log fail: parseValidateState:"+
			parseValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRsyncLogParseValidateStateEndDb(): CommitSession fail:"+
			parseValidateState+","+state+",  labRpkiSyncLogId:", labRpkiSyncLogId, err)
		return err
	}
	return nil
}
