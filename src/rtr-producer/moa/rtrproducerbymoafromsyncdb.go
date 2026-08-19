package moa

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func getAllMoasDb() ([]model.MoaToRtrFullLog, error) {
	// get lastest syncLogFile.Id
	moaToRtrFullLogs := make([]model.MoaToRtrFullLog, 0)
	sql := `select 	a.id as moaId, a.ipv6MappingPrefix, a.ipv4Prefixes,
					a.syncLogId,a.syncLogFileId from lab_rpki_moa a
		 	order by a.id `
	err := xormdb.XormEngine.SQL(sql).Find(&moaToRtrFullLogs)
	if err != nil {
		belogs.Error("getAllMoasDb(): find fail:", err)
		return nil, err
	}
	belogs.Debug("getAllMoasDb(): len(moaToRtrFullLogs):", len(moaToRtrFullLogs))
	return moaToRtrFullLogs, nil
}

func getRtrMoaFullFromRtrFullLogDb(serialNumber uint64) (rtrMoaFulls map[string]model.LabRpkiRtrMoaFull, err error) {
	start := time.Now()
	belogs.Debug("getRtrMoaFullFromRtrFullLogDb():serialNumber:", serialNumber)
	rtrMoaFs := make([]model.LabRpkiRtrMoaFull, 0)
	sql :=
		`select serialNumber,ipv6MappingPrefix,ipv4Prefixes,sourceFrom 
	    from lab_rpki_rtr_moa_full_log 
	    where serialNumber = ? 
		order by id `
	err = xormdb.XormEngine.SQL(sql, serialNumber).Find(&rtrMoaFs)
	if err != nil {
		belogs.Error("getRtrMoaFullFromRtrFullLogDb(): get lab_rpki_rtr_moa_full_log fail: serialNumber: ", serialNumber, err)
		return nil, err
	}
	if len(rtrMoaFs) == 0 {
		belogs.Debug("getRtrMoaFullFromRtrFullLogDb(): len(rtrMoaFs)==0: serialNumber", serialNumber)
		return make(map[string]model.LabRpkiRtrMoaFull, 0), nil
	}
	belogs.Debug("getRtrMoaFullFromRtrFullLogDb():model.LabRpkiRtrMoaFull, serialNumber, len(rtrMoaFs) : ", serialNumber, len(rtrMoaFs))

	rtrMoaFulls = make(map[string]model.LabRpkiRtrMoaFull, len(rtrMoaFs)+50)
	for i := range rtrMoaFs {
		key := rtrMoaFs[i].Ipv6MappingPrefix + "_" + rtrMoaFs[i].Ipv4Prefixes
		rtrMoaFulls[key] = rtrMoaFs[i]
	}
	belogs.Info("getRtrMoaFullFromRtrFullLogDb():map LabRpkiRtrMoaFull, serialNumber:",
		serialNumber, "  , len(rtrMoaFs):", len(rtrMoaFs), "   time(s):", time.Since(start))
	return rtrMoaFulls, nil

}

func insertRtrMoaFullLogFromMoaDb(newSerialNumber uint64, moaToRtrFullLogs []model.MoaToRtrFullLog) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrMoaFullLogFromMoaDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// insert moa into rtr_moa_full_log
	sql := `insert ignore into lab_rpki_rtr_moa_full_log
				(serialNumber,ipv6MappingPrefix,ipv4Prefixes,
					sourceFrom) values
				(?,?,?,    ?)`
	sourceFrom := model.LabRpkiRtrSourceFrom{
		Source: "sync",
	}
	belogs.Debug("insertRtrMoaFullLogFromMoaDb(): will insert lab_rpki_rtr_moa_full_log from moaToRtrFullLogs, len(moaToRtrFullLogs): ", len(moaToRtrFullLogs))
	for i := range moaToRtrFullLogs {
		sourceFrom.SyncLogId = moaToRtrFullLogs[i].SyncLogId
		sourceFrom.SyncLogFileId = moaToRtrFullLogs[i].SyncLogFileId
		sourceFromJson := jsonutil.MarshalJson(sourceFrom)
		_, err = session.Exec(sql,
			newSerialNumber, moaToRtrFullLogs[i].Ipv6MappingPrefix, moaToRtrFullLogs[i].Ipv4Prefixes,
			sourceFromJson)
		if err != nil {
			belogs.Error("insertRtrMoaFullLogFromMoaDb():insert into lab_rpki_rtr_moa_full_log ipv4 from moa fail:",
				jsonutil.MarshalJson(moaToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrMoaFullLogFromMoaDb(): insert into lab_rpki_rtr_moa_full_log ipv4 fail: ", err)
		}

	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrMoaFullLogFromMoaDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrMoaFullLogFromMoaDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrMoaFullLogFromMoaDb(): CommitSession ok, len(moaToRtrFullLogs): ", len(moaToRtrFullLogs), "   time(s):", time.Since(start))
	return nil
}

func updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(newSerialNumberModel *rtrcommon.SerialNumberModel,
	rtrMoaIncrementals []model.LabRpkiRtrMoaIncremental) (err error) {
	start := time.Now()
	belogs.Debug("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrMoaIncrementals):", len(rtrMoaIncrementals))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// serialnumber/rtrmoafull/rtrmoaincr should in one session
	// insert new serial number
	err = rtrcommon.InsertSerialNumberDb(session, newSerialNumberModel, start)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():InsertSerialNumberDb fail,newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():InsertSerialNumberDb fail:", err)
	}
	belogs.Debug("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():InsertSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// delete and insert into lab_rpki_rtr_moa_full
	sql := `delete from lab_rpki_rtr_moa_full`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():delete lab_rpki_rtr_moa_full fail:", err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():delete lab_rpki_rtr_moa_full fail:", err)
	}
	belogs.Debug("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():delete lab_rpki_rtr_moa_full, time(s):", time.Since(start))

	// insert rtr_moa_full from rtr_full_moa_log
	sql = `
	insert ignore into lab_rpki_rtr_moa_full 
		  (serialNumber, ipv6MappingPrefix,ipv4Prefixes,
		   sourceFrom ) 
	select serialNumber, ipv6MappingPrefix,ipv4Prefixes,
	        sourceFrom 
	from lab_rpki_rtr_moa_full_log where serialNumber=? order by id`
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():insert into lab_rpki_rtr_moa_full from lab_rpki_rtr_moa_full_log fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():insert into lab_rpki_rtr_moa_full from lab_rpki_rtr_moa_full_log fail: ", err)
	}
	belogs.Debug("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():insert into lab_rpki_rtr_moa_full from lab_rpki_rtr_moa_full_log , time(s):", time.Since(start))

	// insert into lab_rpki_rtr_moa_incremental
	sql = `insert ignore into lab_rpki_rtr_moa_incremental
		(serialNumber,style,ipv6MappingPrefix,ipv4Prefixes,   sourceFrom) values
		(?,?,?,?,  ?)`
	for i := range rtrMoaIncrementals {
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, rtrMoaIncrementals[i].Style,
			rtrMoaIncrementals[i].Ipv6MappingPrefix, rtrMoaIncrementals[i].Ipv4Prefixes,
			rtrMoaIncrementals[i].SourceFrom)
		if err != nil {
			belogs.Error("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():insert into lab_rpki_rtr_moa_incremental fail: newSerialNumber:",
				jsonutil.MarshalJson(newSerialNumberModel), jsonutil.MarshalJson(rtrMoaIncrementals[i]), err)
			return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():insert into lab_rpki_rtr_moa_incremental fail: ", err)
		}
	}
	belogs.Debug("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb():insert into lab_rpki_rtr_moa_incremental, time(s):", time.Since(start))

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(): CommitSession fail: ", err)
	}

	belogs.Info("updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(): CommitSession ok: newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrMoaIncrementals):", len(rtrMoaIncrementals), "   time(s):", time.Since(start))
	return nil
}
