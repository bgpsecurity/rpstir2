package rtrproducer

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func getAllRoasDb() (roaToRtrFullLogs []model.RoaToRtrFullLog, err error) {
	// get lastest syncLogFile.Id

	sql :=
		`SELECT
		r.id AS roaId,
		r.asn AS asn,
		substring_index( i.addressPrefix, '/', 1 ) AS address,
		substring_index( i.addressPrefix, '/', -1 ) AS prefixLength,
		i.maxLength AS maxLength,
		r.syncLogId AS syncLogId,
		r.syncLogFileId AS syncLogFileId 
	FROM
		( lab_rpki_roa r , lab_rpki_roa_ipaddress i ) 
	WHERE
		( i.roaId = r.id and r.state->'$.state' in ('valid','warning')  ) 
	ORDER BY
	    address,prefixLength desc,maxLength desc,r.asn,i.id		 `
	err = xormdb.XormEngine.SQL(sql).Find(&roaToRtrFullLogs)
	if err != nil {
		belogs.Error("getAllRoasDb(): find fail:", err)
		return roaToRtrFullLogs, err
	}
	belogs.Debug("getAllRoasDb(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs))

	return roaToRtrFullLogs, nil
}

func updateRtrFullLogFromRoaAndSlurmDb(newSerialNumberModel SerialNumberModel, roaToRtrFullLogs []model.RoaToRtrFullLog,
	effectSlurmToRtrFullLogs []model.EffectSlurmToRtrFullLog) (err error) {
	//when both  len are 0, return nil
	if len(roaToRtrFullLogs) == 0 && len(effectSlurmToRtrFullLogs) == 0 {
		belogs.Info("updateRtrFullLogFromRoaAndSlurmDb():roa and slurm all are empty")
		return nil
	}
	start := time.Now()

	err = insertRtrFullLogFromRoaDb(newSerialNumberModel.SerialNumber, roaToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrFullLogFromRoaAndSlurmDb():insertRtrFullLogFromRoaDb fail:", err)
		return err
	}
	belogs.Info("updateRtrFullLogFromRoaAndSlurmDb():insertRtrFullLogFromRoaDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(roaToRtrFullLogs):", len(roaToRtrFullLogs), "  time(s):", time.Since(start))

	err = updateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full_log", newSerialNumberModel.SerialNumber, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrFullLogFromRoaAndSlurmDb(): updateRtrFullOrFullLogFromSlurmDb fail:", err)
		return err
	}
	belogs.Info("updateRtrFullLogFromRoaAndSlurmDb():updateRtrFullOrFullLogFromSlurmDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Since(start))
	return nil
}

func updateSerailNumberAndRtrFullAndRtrIncrementalDb(newSerialNumberModel SerialNumberModel,
	rtrIncrementals []model.LabRpkiRtrIncremental) (err error) {
	start := time.Now()
	belogs.Debug("updateSerailNumberAndRtrFullAndRtrIncrementalDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrIncrementals):", len(rtrIncrementals))

	session, err := xormdb.NewSession()
	defer session.Close()

	// serialnumber/rtrfull/rtrincr should in one session
	// insert new serial number
	sql := ` insert into lab_rpki_rtr_serial_number(
		serialNumber,globalSerialNumber,subpartSerialNumber, createTime)
		 values(?,?,?,?)`
	_, err = session.Exec(sql,
		newSerialNumberModel.SerialNumber, newSerialNumberModel.GlobalSerialNumber,
		newSerialNumberModel.SubpartSerialNumber, start)
	if err != nil {
		belogs.Error("updateSerailNumberAndRtrFullAndRtrIncrementalDb():insert into lab_rpki_rtr_serial_number fail:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return err
	}

	// delete and insert into lab_rpki_rtr_full
	sql = `delete from lab_rpki_rtr_full`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():delete lab_rpki_rtr_full fail:", err)
		return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrFullAndRtrIncrementalDb():delete lab_rpki_rtr_full fail: ", err)
	}

	// insert rtr_full from rtr_full_log
	sql = `
	insert ignore into lab_rpki_rtr_full 
		(serialNumber, asn ,address, 
		prefixLength,maxLength, sourceFrom ) 
	select serialNumber, asn ,address, 
		prefixLength,maxLength, sourceFrom 
	from lab_rpki_rtr_full_log where serialNumber=? order by id`
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateSerailNumberAndRtrFullAndRtrIncrementalDb():insert into lab_rpki_rtr_full from lab_rpki_rtr_full_log fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrFullAndRtrIncrementalDb():insert into lab_rpki_rtr_full from lab_rpki_rtr_full_log fail: ", err)
	}

	// insert into lab_rpki_rtr_incremental
	sql = `insert ignore into lab_rpki_rtr_incremental
		(serialNumber,style,asn,address,   prefixLength,maxLength, sourceFrom) values
		(?,?,?,?,  ?,?,?)`
	for i := range rtrIncrementals {
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, rtrIncrementals[i].Style, rtrIncrementals[i].Asn, rtrIncrementals[i].Address,
			rtrIncrementals[i].PrefixLength, rtrIncrementals[i].MaxLength, rtrIncrementals[i].SourceFrom)
		if err != nil {
			belogs.Error("updateSerailNumberAndRtrFullAndRtrIncrementalDb():insert into lab_rpki_rtr_incremental fail: newSerialNumber:",
				jsonutil.MarshalJson(newSerialNumberModel), jsonutil.MarshalJson(rtrIncrementals[i]), err)
			return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrFullAndRtrIncrementalDb():insert into lab_rpki_rtr_full from lab_rpki_rtr_full_log fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateSerailNumberAndRtrFullAndRtrIncrementalDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrFullAndRtrIncrementalDb(): CommitSession fail: ", err)
	}

	belogs.Info("updateSerailNumberAndRtrFullAndRtrIncrementalDb(): CommitSession ok: newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrIncrementals):", len(rtrIncrementals), "   time(s):", time.Since(start))
	return nil
}

func getRtrFullFromRtrFullLogDb(serialNumber uint64) (rtrFulls map[string]model.LabRpkiRtrFull, err error) {
	start := time.Now()
	belogs.Debug("getRtrFullFromRtrFullLogDb():serialNumber:", serialNumber)
	rtrFs := make([]model.LabRpkiRtrFull, 0)
	/* sql :=
		`select asn, address, prefixLength, maxlength, sourceFrom
	    from lab_rpki_rtr_full_log
	    where serialNumber = ?
		order by id `
	*/
	err = xormdb.XormEngine.Table("lab_rpki_rtr_full_log").
		Cols("asn, address, prefixLength, maxlength, sourceFrom").
		Where("serialNumber = ?", serialNumber).OrderBy("id").Find(&rtrFs)
	if err != nil {
		belogs.Error("getRtrFullFromRtrFullLogDb(): find fail: serialNumber: ", serialNumber, err)
		return nil, err
	}
	if len(rtrFs) == 0 {
		belogs.Debug("getRtrFullFromRtrFullLogDb(): len(rtrFs)==0: serialNumber", serialNumber)
		return rtrFulls, nil
	}
	belogs.Debug("getRtrFullFromRtrFullLogDb():model.LabRpkiRtrFull, serialNumber, len(rtrFs) : ", serialNumber, len(rtrFs))

	rtrFulls = make(map[string]model.LabRpkiRtrFull, len(rtrFs)+50)
	for i := range rtrFs {
		key := convert.ToString(rtrFs[i].Asn) + "_" + rtrFs[i].Address + "_" +
			convert.ToString(rtrFs[i].PrefixLength) + "_" + convert.ToString(rtrFs[i].MaxLength)
		rtrFulls[key] = rtrFs[i]
	}
	belogs.Info("getRtrFullFromRtrFullLogDb():map LabRpkiRtrFull, serialNumber, len(rtrFulls):", serialNumber, len(rtrFulls), "   time(s):", time.Since(start))
	return rtrFulls, nil

}

func insertRtrFullLogFromRoaDb(newSerialNumber uint64, roaToRtrFullLogs []model.RoaToRtrFullLog) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	// insert roa into rtr_full_log
	sql := `insert  into lab_rpki_rtr_full_log
				(serialNumber,asn,address,prefixLength, maxLength,sourceFrom) values
				(?,?,?,  ?,?,?)`
	sourceFrom := model.LabRpkiRtrSourceFrom{
		Source: "sync",
	}
	belogs.Debug("insertRtrFullLogFromRoaDb(): will insert lab_rpki_rtr_full_log from roaToRtrFullLogs, len(roaToRtrFullLogs): ", len(roaToRtrFullLogs))
	for i := range roaToRtrFullLogs {
		sourceFrom.SyncLogId = roaToRtrFullLogs[i].SyncLogId
		sourceFrom.SyncLogFileId = roaToRtrFullLogs[i].SyncLogFileId
		sourceFromJson := jsonutil.MarshalJson(sourceFrom)

		_, err = session.Exec(sql,
			newSerialNumber, roaToRtrFullLogs[i].Asn, roaToRtrFullLogs[i].Address,
			roaToRtrFullLogs[i].PrefixLength, roaToRtrFullLogs[i].MaxLength, sourceFromJson)
		if err != nil {
			belogs.Error("insertRtrFullLogFromRoaDb():insert into lab_rpki_rtr_full_log from roa fail:",
				jsonutil.MarshalJson(roaToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrFullLogFromRoaDb(): insert into lab_rpki_rtr_full_log fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrFullLogFromRoaDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrFullLogFromRoaDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrFullLogFromRoaDb(): CommitSession ok, len(roaToRtrFullLogs): ", len(roaToRtrFullLogs), "   time(s):", time.Since(start))
	return nil
}
