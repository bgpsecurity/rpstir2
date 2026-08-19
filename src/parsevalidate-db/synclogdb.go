package parsevalidatedb

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

// state: parseValidating;
func UpdateSyncLogParseValidateStartDb(state string) (labRpkiSyncLogId uint64, err error) {
	belogs.Debug("UpdateSyncLogParseValidateStartDb():  state:", state)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateSyncLogParseValidateStartDb(): NewSession fail:", err)
		return 0, err
	}
	defer session.Close()

	var id int64
	_, err = session.SQL("SELECT MAX(id) AS id FROM lab_rpki_sync_log").Get(&id) //_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateSyncLogParseValidateStartDb(): update sync_log fail: state:"+state, err)
	}
	syncLogParseValidateState := model.SyncLogParseValidateState{
		StartTime: time.Now(),
	}
	parseValidateState := jsonutil.MarshalJson(syncLogParseValidateState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set parseValidateState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, parseValidateState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateSyncLogParseValidateStartDb(): UPDATE sync_log fail: parseValidateState:"+
			parseValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateSyncLogParseValidateStartDb(): CommitSession fail:"+
			parseValidateState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	return uint64(id), nil
}
func UpdateSyncLogParseValidateStateEndDb(labRpkiSyncLogId uint64, state string,
	parseFailFiles []string) (err error) {
	belogs.Debug("UpdateSyncLogParseValidateStateEndDb(): labRpkiSyncLogId:", labRpkiSyncLogId)
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateSyncLogParseValidateStateEndDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	// get current parseValidateState, the set new value
	var parseValidateState string
	_, err = session.Table("lab_rpki_sync_log").Cols("parseValidateState").Where("id = ?", labRpkiSyncLogId).Get(&parseValidateState)
	if err != nil {
		belogs.Error("UpdateSyncLogParseValidateStateEndDb(): lab_rpki_sync_log Get parseValidateState :", labRpkiSyncLogId, err)
		return err
	}
	belogs.Debug("UpdateSyncLogParseValidateStateEndDb(): get parseValidateState, labRpkiSyncLogId:", labRpkiSyncLogId, "  parseValidateState:", parseValidateState)

	syncLogParseValidateState := model.SyncLogParseValidateState{}
	jsonutil.UnmarshalJson(parseValidateState, &syncLogParseValidateState)
	syncLogParseValidateState.EndTime = time.Now()
	syncLogParseValidateState.ParseFailFiles = parseFailFiles
	parseValidateState = jsonutil.MarshalJson(syncLogParseValidateState)
	belogs.Debug("UpdateSyncLogParseValidateStateEndDb():parseValidateState:", parseValidateState, "  labRpkiSyncLogId:", labRpkiSyncLogId)

	sqlStr := `UPDATE lab_rpki_sync_log set parseValidateState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, parseValidateState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("UpdateSyncLogParseValidateStateEndDb(): lab_rpki_sync_log UPDATE :", labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "UpdateSyncLogParseValidateStateEndDb(): UPDATE sync_log fail: parseValidateState:"+
			parseValidateState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	belogs.Debug("UpdateSyncLogParseValidateStateEndDb(): state:", state, "  parseValidateState:", parseValidateState, "  labRpkiSyncLogId:", labRpkiSyncLogId)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateSyncLogParseValidateStateEndDb(): CommitSession fail:"+
			parseValidateState+","+state+",  labRpkiSyncLogId:", labRpkiSyncLogId, err)
		return err
	}
	belogs.Debug("UpdateSyncLogParseValidateStateEndDb():ok, state:", state, "  parseValidateState:", parseValidateState, "  labRpkiSyncLogId:", labRpkiSyncLogId)
	return nil
}
