package rrdp

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

func GetLastSyncRrdpLogsDb() (syncRrdpLogs map[string]model.LabRpkiSyncRrdpLog, err error) {
	start := time.Now()
	/*
		sql := `select c.id,  c.syncLogId,  c.notifyUrl,  c.sessionId,  c.lastSerial,
				  c.curSerial,  c.rrdpTime,  c.rrdpType
			   from lab_rpki_sync_rrdp_log c
			   where c.syncLogId = (select syncLogId from lab_rpki_sync_rrdp_log cc
									where cc.notifyUrl = c.notifyUrl order by cc.syncLogId desc limit 1)
			   order by c.notifyUrl `
	*/
	sql := `select c.id,  c.syncLogId,  c.notifyUrl,  c.sessionId,  c.lastSerial, 
	           c.curSerial,  c.rrdpTime,  c.rrdpType 
	        from lab_rpki_sync_rrdp_log c , lab_rpki_sync_rrdp_log_maxid_view v 
		    where c.id = v.maxid 
			order by c.notifyUrl `
	rrdps := make([]model.LabRpkiSyncRrdpLog, 0)
	err = xormdb.XormEngine.SQL(sql).Find(&rrdps)
	if err != nil {
		belogs.Error("GetLastSyncRrdpLogsDb(): find fail:", err)
		return nil, err
	}
	belogs.Info("GetLastSyncRrdpLogsDb(): len(rrdps):", len(rrdps), " ,  time(s):", time.Since(start))

	syncRrdpLogs = make(map[string]model.LabRpkiSyncRrdpLog)
	for i := range rrdps {
		syncRrdpLogs[rrdps[i].NotifyUrl] = rrdps[i]
		belogs.Info("GetLastSyncRrdpLogsDb(): rrdp :", jsonutil.MarshalJson(rrdps[i]))
	}
	belogs.Info("GetLastSyncRrdpLogsDb(): len(syncRrdpLogs):", len(syncRrdpLogs),
		" ,  time(s):", time.Since(start))
	return syncRrdpLogs, nil
}

func GetLastSyncRrdpLogDb(rrdpUrl string) (syncRrdpLog model.LabRpkiSyncRrdpLog, has bool, err error) {
	start := time.Now()
	sql := `select c.id,  c.syncLogId,  c.notifyUrl,  c.sessionId,  c.lastSerial,  
			  c.curSerial,  c.rrdpTime,  c.rrdpType  
		   	from lab_rpki_sync_rrdp_log c   
		   	where c.syncLogId = (select syncLogId from lab_rpki_sync_rrdp_log cc  
								where cc.notifyUrl = c.notifyUrl order by cc.syncLogId desc limit 1)  
			and c.notifyUrl = ? 	
			order by c.notifyUrl `
	has, err = xormdb.XormEngine.SQL(sql, rrdpUrl).Get(&syncRrdpLog)
	if err != nil {
		belogs.Error("getLastSyncRrdpLogDb(): find fail:", rrdpUrl, err)
		return syncRrdpLog, false, err
	}
	belogs.Info("getLastSyncRrdpLogDb(): syncRrdpLog:", jsonutil.MarshalJson(syncRrdpLog),
		" ,  time(s):", time.Since(start))
	return syncRrdpLog, has, nil
}

/*
func GetLastSyncRrdpLog(notifyUrl string, sessionId string, minSerial, maxSerial uint64) (has bool, syncRrdpLog model.LabRpkiSyncRrdpLog, err error) {
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
	has, err = xormdb.XormEngine.SQL(sql, notifyUrl, sessionId, minSerial, maxSerial).Get(&syncRrdpLog)
	if err != nil {
		belogs.Error("GetLastSyncRrdpLog(): find fail:", notifyUrl, err)
		return false, syncRrdpLog, err
	}
	belogs.Debug("GetLastSyncRrdpLog(): has, syncRrdpLog:", has, jsonutil.MarshalJson(syncRrdpLog))

	return has, syncRrdpLog, nil
}
*/
func insertSyncRrdpLogDb(
	session *xorm.Session, syncLogId uint64, snapshotDeltaResult *SnapshotDeltaResult) (err error) {

	//then rrdpType is snapshot, lastSerial is 0(will set null in mysql), and default rrdpType is snapshot
	belogs.Debug("insertSyncRrdpLogDb(): syncLogId ,  snapshotDeltaResult :", syncLogId, jsonutil.MarshalJson(snapshotDeltaResult))

	//lab_rpki_sync_log
	sqlStr := `INSERT lab_rpki_sync_rrdp_log(syncLogId,  notifyUrl,  sessionId,  
				lastSerial,	  curSerial,  
				rrdpTime,  rrdpType, snapshotOrDeltaUrl)
				VALUES(?,?,?,   ?,?,    ?,?,?)`
	_, err = session.Exec(sqlStr, syncLogId, snapshotDeltaResult.NotifyUrl, snapshotDeltaResult.SessionId,
		xormdb.SqlNullInt(int64(snapshotDeltaResult.LastSerial)), snapshotDeltaResult.Serial,
		snapshotDeltaResult.RrdpTime, snapshotDeltaResult.RrdpType, snapshotDeltaResult.SnapshotOrDeltaUrl)
	if err != nil {
		belogs.Error("insertSyncRrdpLogDb(): INSERT lab_rpki_sync_rrdp_log fail:",
			syncLogId, jsonutil.MarshalJson(snapshotDeltaResult), err)
		return err
	}

	return nil
}
