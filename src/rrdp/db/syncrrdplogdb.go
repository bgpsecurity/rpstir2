package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

func GetLastSyncRrdpLog(notifyUrl string, sessionId string, minSerail, maxSerail uint64) (has bool, syncRrdpLog model.LabRpkiSyncRrdpLog, err error) {
	sql :=
		`SELECT
		  id,  syncLogId,  notifyUrl,  sessionId,  lastSerial,
		  curSerial,  rrdpTime,  rrdpType
		FROM
			lab_rpki_sync_rrdp_log 
		WHERE
			syncLogId = ( select max(syncLogId) from lab_rpki_sync_rrdp_log ) 
			and notifyUrl=? 
			and sessionId=? 
			and curSerial >= ? and curSerial <= ?			
		ORDER BY
			id	`
	has, err = xormdb.XormEngine.Sql(sql, notifyUrl, sessionId, minSerail, maxSerail).Get(&syncRrdpLog)
	if err != nil {
		belogs.Error("GetLastSyncRrdpLog(): find fail:", notifyUrl, err)
		return false, syncRrdpLog, err
	}
	belogs.Debug("GetLastSyncRrdpLog(): has, syncRrdpLog:", has, jsonutil.MarshalJson(syncRrdpLog))

	return has, syncRrdpLog, nil
}

func InsertSyncRrdpLog(
	has bool, syncLogId uint64, notifyUrl string,
	syncRrdpLog *model.LabRpkiSyncRrdpLog, notificationModel *rrdputil.NotificationModel) (err error) {

	session, err := xormdb.NewSession()
	defer session.Close()
	//when "has" is false, it is snapshot,
	//then default lastSerial is -1(will set null in mysql), and default rrdpType is snapshot
	lastSerial := int64(-1)
	rrdpType := "snapshot"
	if has {
		lastSerial = int64(syncRrdpLog.CurSerial)
		rrdpType = "delta"
	}
	belogs.Debug("InsertSyncRrdpLog():has, syncLogId, notifyUrl,notificationModel.SessionId, "+
		"lastSerial, notificationModel.Serial, rrdpType:", has, syncLogId, notifyUrl,
		notificationModel.SessionId, lastSerial, notificationModel.Serial, rrdpType)

	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_rrdp_log(syncLogId,  notifyUrl,  sessionId,  
				lastSerial,	  curSerial,  rrdpTime,  rrdpType)
				VALUES(?,?,?,   ?,?,?,?)`
	_, err = session.Exec(sqlStr, syncLogId, notifyUrl, notificationModel.SessionId,
		xormdb.SqlNullInt(lastSerial), notificationModel.Serial, time.Now(), rrdpType)
	if err != nil {
		belogs.Error("InsertSyncRrdpLog(): INSERT lab_rpki_sync_rrdp_log fail:",
			syncRrdpLog, notifyUrl, notificationModel.SessionId,
			lastSerial, notificationModel.Serial, time.Now(), rrdpType, err)
		return xormdb.RollbackAndLogError(session, "INSERT lab_rpki_sync_rrdp_log fail:", err)
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "InsertSyncRrdpLog(): CommitSession fail:", err)
	}

	return nil
}
