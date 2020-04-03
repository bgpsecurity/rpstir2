package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

// state: rrdping;  style: rrdp
func InsertRsyncLogRrdpStateStart(state, syncStyle string) (logId uint64, err error) {
	syncLogRrdpState := model.SyncLogRrdpState{
		StartTime: time.Now(),
	}
	rrdpState := jsonutil.MarshalJson(syncLogRrdpState)

	session, err := xormdb.NewSession()
	defer session.Close()
	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_log(rrdpState, state,syncStyle)
					VALUES(?,?,?)`
	res, err := session.Exec(sqlStr, rrdpState, state, syncStyle)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertRsyncLogRrdpStateStart(): INSERT lab_rpki_sync_log fail:"+rrdpState+","+state+","+syncStyle, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertRsyncLogRrdpStateStart(): LastInsertId fail:"+rrdpState+","+state+","+syncStyle, err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertRsyncLogRrdpStateStart(): CommitSession fail:"+rrdpState+","+state+","+syncStyle, err)

	}

	belogs.Debug("InsertRsyncLogRrdpStateStart():LastInsertId:", id)
	return uint64(id), nil
}

func UpdateRsyncLogRrdpStateEnd(labRpkiSyncLogId uint64, state string) (err error) {
	belogs.Debug("UpdateRsyncLogRrdpStateEnd():labRpkiSyncLogId:", labRpkiSyncLogId, "   state:", state)

	session, err := xormdb.NewSession()
	defer session.Close()

	// get current rrdpState, the set new value
	var rrdpState string
	_, err = session.Table("lab_rpki_sync_log").Cols("rrdpState").Where("id = ?", labRpkiSyncLogId).Get(&rrdpState)
	if err != nil {
		belogs.Error("UpdateRsyncLogRrdpStateEnd(): lab_rpki_sync_log Get rrdpState :", labRpkiSyncLogId, err)
		return err
	}
	syncLogRrdpState := model.SyncLogRrdpState{}
	jsonutil.UnmarshalJson(rrdpState, &syncLogRrdpState)
	syncLogRrdpState.EndTime = time.Now()
	rrdpState = jsonutil.MarshalJson(syncLogRrdpState)
	belogs.Debug("UpdateRsyncLogRrdpStateEnd():rrdpState:", rrdpState)

	sqlStr := `UPDATE lab_rpki_sync_log set rrdpState=?, state=? where id=? `
	_, err = session.Exec(sqlStr, rrdpState, state, labRpkiSyncLogId)
	if err != nil {
		belogs.Error("UpdateRsyncLogRrdpStateEnd(): UPDATE lab_rpki_sync_log fail : rrdpState: "+
			rrdpState+"   state:"+state+"    labRpkiSyncLogId:"+convert.ToString(labRpkiSyncLogId), err)
		return err
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "ProcessRrdpSnapshot(): CommitSession fail:", err)
	}
	return nil
}
