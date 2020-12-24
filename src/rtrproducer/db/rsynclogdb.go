package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
)

// state: chainValidating;
func UpdateRsyncLogRtrStateStart(state string) (labRpkiSyncLogId uint64, err error) {
	belogs.Debug("UpdateRsyncLogRtrStateStart():  state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	var id int64
	_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogRtrStateStart(): update lab_rpki_sync_log fail: state:"+state, err)
	}
	syncLogRtrState := model.SyncLogRtrState{
		StartTime: time.Now(),
	}
	rtrState := jsonutil.MarshalJson(syncLogRtrState)

	//lab_rpki_sync_log
	sqlStr := `UPDATE lab_rpki_sync_log set rtrState=?, state=? where id=?`
	_, err = session.Exec(sqlStr, rtrState, state, id)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogRtrStateStart(): UPDATE lab_rpki_sync_log fail: rtrState:"+
			rtrState+",   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(id), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "UpdateRsyncLogRtrStateStart(): CommitSession fail:"+
			rtrState+","+state+",  labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}
	return uint64(id), nil
}

func UpdateRsyncLogRtrStateEnd(session *xorm.Session, labRpkiSyncLogId uint64, state string) (err error) {
	// get current rtrState, the set new value

	var rtrState string
	_, err = session.Table("lab_rpki_sync_log").Cols("rtrState").Where("id = ?", labRpkiSyncLogId).Get(&rtrState)
	if err != nil {
		belogs.Error("UpdateRsyncLogRtrStateEnd(): lab_rpki_sync_log Get rtrState :", labRpkiSyncLogId, err)
		return err
	}
	syncLogRtrState := model.SyncLogRtrState{}
	jsonutil.UnmarshalJson(rtrState, &syncLogRtrState)
	syncLogRtrState.EndTime = time.Now()
	rtrState = jsonutil.MarshalJson(syncLogRtrState)
	belogs.Debug("UpdateRsyncLogRtrStateEnd():syncLogRtrState:", rtrState)

	sqlStr := `UPDATE lab_rpki_sync_log set rtrState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, rtrState, state, labRpkiSyncLogId)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRsyncLogRtrStateEnd(): UPDATE lab_rpki_sync_log fail : rtrState: "+
			rtrState+"   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
	}

	return nil
}
