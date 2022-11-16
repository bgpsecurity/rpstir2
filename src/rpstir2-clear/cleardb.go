package clear

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func clearSyncLogFileDb() (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	// get max labRpkiSyncLogId
	var labRpkiSyncLogId int64
	_, err = session.Table("lab_rpki_sync_log").Select("max(id)").Get(&labRpkiSyncLogId)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): max labRpkiSyncLogId fail: ", err)
	}

	deleteLabRpkiSyncLogId := int64(labRpkiSyncLogId - 24)
	belogs.Info("clearSyncLogFileDb():labRpkiSyncLogId:", labRpkiSyncLogId, "   deleteLabRpkiSyncLogId:", deleteLabRpkiSyncLogId)
	if deleteLabRpkiSyncLogId <= 0 {
		return
	}

	// mft
	err = clearSyncLogFileImplDb(session, deleteLabRpkiSyncLogId, "lab_rpki_mft", "mft")
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): lab_rpki_mft fail:", err)
	}

	// roa
	err = clearSyncLogFileImplDb(session, deleteLabRpkiSyncLogId, "lab_rpki_roa", "roa")
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): lab_rpki_roa fail:", err)
	}

	// cer
	err = clearSyncLogFileImplDb(session, deleteLabRpkiSyncLogId, "lab_rpki_cer", "cer")
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): lab_rpki_cer fail:", err)
	}

	// crl
	err = clearSyncLogFileImplDb(session, deleteLabRpkiSyncLogId, "lab_rpki_crl", "crl")
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): lab_rpki_crl fail:", err)
	}

	// optimize
	sql := `optimize  table  lab_rpki_sync_log_file `
	_, err = session.Exec(sql)
	belogs.Debug("clearSyncLogFileDb(): optimize ")
	if err != nil {
		belogs.Error("clearSyncLogFileDb():optimize fail:", err)
		return err
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): CommitSession fail:", err)
	}
	belogs.Info("clearSyncLogFileDb(): end, time(s):", time.Since(start))
	return nil
}

//  tableName: lab_rpki_mft
// fileType: mft
func clearSyncLogFileImplDb(session *xorm.Session, deleteLabRpkiSyncLogId int64, tableName string, fileType string) (err error) {
	sql := `delete from lab_rpki_sync_log_file f 
			where f.state->'$.updateCertTable' = 'finished' 
				and f.id not in (select distinct(syncLogFileId) from  ` + tableName + ` s where s.syncLogId < ?  )	 
				and f.fileType= ? and f.syncLogId < ? `
	affected, err := session.Exec(sql, deleteLabRpkiSyncLogId, fileType, deleteLabRpkiSyncLogId)
	if err != nil {
		belogs.Error("clearSyncLogFileImplDb(): delete lab_rpki_sync_log_file fail : sql: ",
			sql, deleteLabRpkiSyncLogId, tableName, fileType, err)
		return err
	}
	deleteRows, err := affected.RowsAffected()
	if err != nil {
		belogs.Error("clearSyncLogFileImplDb():delete deleteRows, RowsAffected :", affected, tableName, err)
		return err
	}
	belogs.Info("clearSyncLogFileImplDb(): end, deleteLabRpkiSyncLogId:", deleteLabRpkiSyncLogId,
		"      deleteRows:", deleteRows, tableName)
	return nil
}

// tableName: lab_rpki_rtr_full_log/lab_rpki_rtr_incremental
func clearRtrFullLogRtrIncremet(tableName string, serialNumber int) (err error) {
	belogs.Debug("clearRtrFullLogRtrIncremet():tableName:", tableName, " ,   serialNumber:", serialNumber)

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	// delete too old
	belogs.Debug("clearRtrFullLogRtrIncremet():will delete too old serialNumber in "+tableName+" ,serialNumber:", serialNumber)
	sql := `delete from ` + tableName + ` where serialNumber < ? `
	affected, err := session.Exec(sql, serialNumber)
	belogs.Debug("clearRtrFullLogRtrIncremet():delete "+tableName+" serialNumber:", serialNumber, "   affected:", affected)
	if err != nil {
		belogs.Error("clearRtrFullLogRtrIncremet():delete serialNumber:", serialNumber, err)
		return err
	}
	deleteRows, err := affected.RowsAffected()
	if err != nil {
		belogs.Error("clearRtrFullLogRtrIncremet():delete serialNumber, RowsAffected :", affected, err)
		return err
	}
	if deleteRows > 10000 {
		sql = `optimize  table  ` + tableName
		_, err = session.Exec(sql)
		belogs.Debug("clearRtrFullLogRtrIncremet():affected > 10000,  optimize "+tableName+", deleteRows:", deleteRows)
		if err != nil {
			belogs.Error("clearRtrFullLogRtrIncremet():optimize "+tableName+":", err)
			return err
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "clearSyncLogFileDb(): CommitSession fail:", err)
	}
	belogs.Info("clearRtrFullLogRtrIncremet(): end, serialNumber:", serialNumber,
		"   deleteRows:", deleteRows, "    tableName:", tableName, "  time(s):", time.Since(start))

	return nil
}

func getMaxSerialNumberDb() (serialNumber int, has bool, err error) {
	sql := `select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1`
	has, err = xormdb.XormEngine.SQL(sql).Get(&serialNumber)
	if err != nil {
		belogs.Error("getMaxSerialNumberDb():select serialNumber from lab_rpki_rtr_serial_number  fail:", err)
		return serialNumber, false, err
	} else if !has {
		belogs.Debug("getMaxSerialNumberDb():select serialNumber from lab_rpki_rtr_serial_number !has")
		return serialNumber, has, nil
	}
	belogs.Info("getMaxSerialNumberDb():max(serialNumber):", serialNumber)
	return serialNumber, has, nil
}
