package db

import (
	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
	rrdpmodel "rrdp/model"
)

func GetLastSyncRrdpLogsByNotifyUrl() (syncRrdpLogs map[string]model.LabRpkiSyncRrdpLog, err error) {
	sql := `select c.id,  c.syncLogId,  c.notifyUrl,  c.sessionId,  c.lastSerial,  
			  c.curSerial,  c.rrdpTime,  c.rrdpType  
		   from lab_rpki_sync_rrdp_log c   
		   where c.syncLogId = (select max(syncLogId) from lab_rpki_sync_rrdp_log cc  
		                        where cc.notifyUrl = c.notifyUrl)  
		   order by c.notifyUrl `
	rrdps := make([]model.LabRpkiSyncRrdpLog, 0)
	err = xormdb.XormEngine.Sql(sql).Find(&rrdps)
	if err != nil {
		belogs.Error("GetLastSyncRrdpLogsByNotifyUrl(): find fail:", err)
		return nil, err
	}
	belogs.Debug("GetLastSyncRrdpLogsByNotifyUrl(): rrdps:", jsonutil.MarshalJson(rrdps))
	syncRrdpLogs = make(map[string]model.LabRpkiSyncRrdpLog)
	for i := range rrdps {
		syncRrdpLogs[rrdps[i].NotifyUrl] = rrdps[i]
	}
	belogs.Debug("GetLastSyncRrdpLogsByNotifyUrl(): syncRrdpLogs:", jsonutil.MarshalJson(syncRrdpLogs))
	return syncRrdpLogs, nil
}

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
	session *xorm.Session, syncLogId uint64, snapshotDeltaResult *rrdpmodel.SnapshotDeltaResult) (err error) {

	//then rrdpType is snapshot, lastSerial is 0(will set null in mysql), and default rrdpType is snapshot
	belogs.Debug("InsertSyncRrdpLog(): syncLogId ,  snapshotDeltaResult :", syncLogId, jsonutil.MarshalJson(snapshotDeltaResult))

	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_rrdp_log(syncLogId,  notifyUrl,  sessionId,  
				lastSerial,	  curSerial,  rrdpTime,  rrdpType)
				VALUES(?,?,?,   ?,?,?,?)`
	_, err = session.Exec(sqlStr, syncLogId, snapshotDeltaResult.NotifyUrl, snapshotDeltaResult.SessionId,
		xormdb.SqlNullInt(int64(snapshotDeltaResult.LastSerial)), snapshotDeltaResult.Serial, snapshotDeltaResult.RrdpTime, snapshotDeltaResult.RrdpType)
	if err != nil {
		belogs.Error("InsertSyncRrdpLog(): INSERT lab_rpki_sync_rrdp_log fail:",
			syncLogId, jsonutil.MarshalJson(snapshotDeltaResult), err)
		return err
	}

	return nil
}
