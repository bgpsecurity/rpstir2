package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

func GetAllRoas() (roaToRtrFullLogs []model.RoaToRtrFullLog, err error) {
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
		i.id		 `
	err = xormdb.XormEngine.Sql(sql).Find(&roaToRtrFullLogs)
	if err != nil {
		belogs.Error("GetAllRoas(): find fail:", err)
		return roaToRtrFullLogs, err
	}
	belogs.Debug("GetAllRoas(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs))

	return roaToRtrFullLogs, nil
}

func GetAllSlurms() (slurmToRtrFullLogs []model.SlurmToRtrFullLog, err error) {
	// get all slurm, not care state->"$.rtr"='notYet' or 'finished'
	sql :=
		`select id as slurmId, style, asn,
			substring_index( addressPrefix, '/', 1 ) AS address,
			substring_index( addressPrefix, '/', -1 ) AS prefixLength,
			maxLength, 
			slurmLogId,
		    slurmLogFileId 
	    from lab_rpki_slurm  
		order by id `
	err = xormdb.XormEngine.Sql(sql).Find(&slurmToRtrFullLogs)
	if err != nil {
		belogs.Error("GetAllSlurms(): find fail:", err)
		return slurmToRtrFullLogs, err
	}
	belogs.Debug("GetAllSlurms(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	return slurmToRtrFullLogs, nil
}

func UpdateRtrFullLog(roaToRtrFullLogs []model.RoaToRtrFullLog,
	slurmToRtrFullLogs []model.SlurmToRtrFullLog) (serialNumber uint32, err error) {

	//when both  len are 0, return nil
	if len(roaToRtrFullLogs) == 0 && len(slurmToRtrFullLogs) == 0 {
		belogs.Debug("UpdateRtrFullLog():roa and slurm all are empty")
		return 0, nil
	}

	session, err := xormdb.NewSession()
	defer session.Close()

	//save to lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = GetMaxSerialNumber()
	if err != nil {
		belogs.Error("UpdateRtrFullLog():lab_rpki_rtr_serial_number LastInsertId fail:", err)
		return 0, err
	}
	serialNumber = serialNumber + 1
	belogs.Debug("UpdateRtrFullLog(): new serialNumber : ", serialNumber)

	// save new serial number
	now := time.Now()
	sql := "insert into lab_rpki_rtr_serial_number(serialNumber, createTime) values(?,?)"
	_, err = session.Exec(sql, serialNumber, now)
	if err != nil {
		belogs.Error("UpdateRtrFullLog():insert into lab_rpki_rtr_serial_number fail:", err)
		return serialNumber, xormdb.RollbackAndLogError(session, "insert new serialnumber fail: ", err)
	}

	// insert roa into rtr_full_log
	sql = `insert   into lab_rpki_rtr_full_log
				(serialNumber,asn,address,prefixLength, maxLength,sourceFrom) values
				(?,?,?,  ?,?,?)`
	sourceFrom := model.LabRpkiRtrSourceFrom{
		Source: "sync",
	}
	for i := range roaToRtrFullLogs {
		sourceFrom.SyncLogId = roaToRtrFullLogs[i].SyncLogId
		sourceFrom.SyncLogFileId = roaToRtrFullLogs[i].SyncLogFileId
		sourceFromJson := jsonutil.MarshalJson(sourceFrom)

		_, err = session.Exec(sql,
			serialNumber, roaToRtrFullLogs[i].Asn, roaToRtrFullLogs[i].Address,
			roaToRtrFullLogs[i].PrefixLength, roaToRtrFullLogs[i].MaxLength, sourceFromJson)
		if err != nil {
			belogs.Error("UpdateRtrFullLog():insert into lab_rpki_rtr_full_log from roa fail:",
				jsonutil.MarshalJson(roaToRtrFullLogs[i]), err)
			return serialNumber, xormdb.RollbackAndLogError(session, "insert into lab_rpki_rtr_full_log from roa fail: ", err)
		}
	}

	// insert into rtr_full_log
	sql = `insert  ignore into lab_rpki_rtr_full_log
				(serialNumber,asn,address,prefixLength, maxLength,sourceFrom) values
				(?,?,?,  ?,?,?)`
	sourceFrom = model.LabRpkiRtrSourceFrom{
		Source: "slurm",
	}
	for i := range slurmToRtrFullLogs {
		sourceFrom.SlurmId = slurmToRtrFullLogs[i].SlurmId
		sourceFrom.SlurmLogId = slurmToRtrFullLogs[i].SlurmLogId
		sourceFrom.SlurmLogFileId = slurmToRtrFullLogs[i].SlurmLogFileId
		sourceFromJson := jsonutil.MarshalJson(sourceFrom)
		belogs.Debug("UpdateRtrFullLog():slurmToRtrFullLogs[i].Style:", slurmToRtrFullLogs[i].SlurmId,
			slurmToRtrFullLogs[i].Style)
		if slurmToRtrFullLogs[i].Style == "prefixAssertions" {
			maxLength := slurmToRtrFullLogs[i].MaxLength
			if maxLength == 0 {
				maxLength = slurmToRtrFullLogs[i].PrefixLength
			}

			affected, err := session.Exec(sql,
				serialNumber, slurmToRtrFullLogs[i].Asn, slurmToRtrFullLogs[i].Address,
				slurmToRtrFullLogs[i].PrefixLength, maxLength, sourceFromJson)
			belogs.Debug("UpdateRtrFullLog():prefixAssertions slurmToRtrFullLogs[i]:",
				jsonutil.MarshalJson(slurmToRtrFullLogs[i]), "   affected:", affected, "    sourceFromJson:", sourceFromJson)
			if err != nil {
				belogs.Error("UpdateRtrFullLog():insert into lab_rpki_rtr_full_log from slurm fail:",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
				return serialNumber, xormdb.RollbackAndLogError(session, "insert into lab_rpki_rtr_full_log from slurm fail: ", err)
			}
		} else if slurmToRtrFullLogs[i].Style == "prefixFilters" {
			labRpkiRtrFullLog := model.LabRpkiRtrFullLog{}
			change := false
			if slurmToRtrFullLogs[i].Asn.Int64 > 0 {
				labRpkiRtrFullLog.Asn = int64(slurmToRtrFullLogs[i].Asn.Int64)
				change = true
			}
			if slurmToRtrFullLogs[i].PrefixLength > 0 {
				labRpkiRtrFullLog.PrefixLength = slurmToRtrFullLogs[i].PrefixLength
				change = true
			}
			if slurmToRtrFullLogs[i].MaxLength > 0 {
				labRpkiRtrFullLog.MaxLength = slurmToRtrFullLogs[i].MaxLength
				change = true
			}
			if len(slurmToRtrFullLogs[i].Address) > 0 {
				labRpkiRtrFullLog.Address = slurmToRtrFullLogs[i].Address
				change = true
			}
			if !change {
				belogs.Error("UpdateRtrFullLog():not found delete condition from slurm, continue to next, :",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]))
				continue
			}
			labRpkiRtrFullLog.SerialNumber = uint64(serialNumber)

			affected, err := session.Delete(&labRpkiRtrFullLog)
			belogs.Debug("UpdateRtrFullLog():prefixFilters slurmToRtrFullLogs[i]:",
				jsonutil.MarshalJson(slurmToRtrFullLogs[i]), "    labRpkiRtrFullLog:", jsonutil.MarshalJson(labRpkiRtrFullLog),
				"   affected:", affected, "    sourceFromJson:", sourceFromJson)
			if err != nil {
				belogs.Error("UpdateRtrFullLog():del lab_rpki_rtr_full_log from slurm fail:",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]), jsonutil.MarshalJson(labRpkiRtrFullLog), err)
				return serialNumber, xormdb.RollbackAndLogError(session, "insert into lab_rpki_rtr_full_log from slurm fail: ", err)
			}
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRtrFullLog(): CommitSession fail :", err)
		return serialNumber, err
	}
	belogs.Debug("UpdateRtrFullLog():CommitSession ok")
	return serialNumber, nil

}
func GetRtrFullFromRtrFullLog(serialNumber uint32) (rtrFulls map[string]model.LabRpkiRtrFull, err error) {

	belogs.Debug("GetRtrFullFromRtrFullLog():serialNumber : ", serialNumber)
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
		belogs.Error("GetRtrFullFromRtrFullLog(): find fail: serialNumber: ", serialNumber, err)
		return nil, err
	}
	if len(rtrFs) == 0 {
		belogs.Debug("GetRtrFullFromRtrFullLog(): len(rtrFs)==0: serialNumber", serialNumber)
		return rtrFulls, nil
	}
	belogs.Debug("GetRtrFullFromRtrFullLog():model.LabRpkiRtrFull,serialNumber, len(rtrFs) : ", serialNumber, len(rtrFs))

	rtrFulls = make(map[string]model.LabRpkiRtrFull, len(rtrFs)+50)
	for i := range rtrFs {
		key := convert.ToString(rtrFs[i].Asn) + "_" + rtrFs[i].Address + "_" +
			convert.ToString(rtrFs[i].PrefixLength) + "_" + convert.ToString(rtrFs[i].MaxLength)
		rtrFulls[key] = rtrFs[i]
	}
	belogs.Debug("GetRtrFullFromRtrFullLog():map LabRpkiRtrFull,serialNumber, len(rtrFulls) : ", serialNumber, len(rtrFulls))
	return rtrFulls, nil

}

func UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd(serialNumber uint32, rtrIncrementals []model.LabRpkiRtrIncremental,
	labRpkiSyncLogId uint64, state string) (err error) {

	session, err := xormdb.NewSession()
	defer session.Close()

	// insert into rtr_full
	sql := `delete from lab_rpki_rtr_full`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd():delete lab_rpki_rtr_full fail:",
			err)
		return xormdb.RollbackAndLogError(session, "delete lab_rpki_rtr_full fail: ", err)
	}
	sql = `
	insert ignore into lab_rpki_rtr_full 
		(serialNumber, asn ,address, 
		prefixLength,maxLength, sourceFrom ) 
	select serialNumber, asn ,address, 
		prefixLength,maxLength, sourceFrom 
	from lab_rpki_rtr_full_log where serialNumber=? order by id`
	_, err = session.Exec(sql, serialNumber)
	if err != nil {
		belogs.Error("UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd():insert into lab_rpki_rtr_full from lab_rpki_rtr_full_log fail: serialNumber:",
			serialNumber, err)
		return xormdb.RollbackAndLogError(session, "insert into lab_rpki_rtr_full from lab_rpki_rtr_full_log fail: ", err)
	}

	sql = `insert ignore into lab_rpki_rtr_incremental
			 (serialNumber,style,asn,address,   prefixLength,maxLength, sourceFrom) values
			 (?,?,?,?,  ?,?,?)`
	for i := range rtrIncrementals {
		_, err = session.Exec(sql,
			serialNumber, rtrIncrementals[i].Style, rtrIncrementals[i].Asn, rtrIncrementals[i].Address,
			rtrIncrementals[i].PrefixLength, rtrIncrementals[i].MaxLength, rtrIncrementals[i].SourceFrom)
		if err != nil {
			belogs.Error("UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd():insert into lab_rpki_rtr_incremental fail: serialNumber:",
				serialNumber, jsonutil.MarshalJson(rtrIncrementals[i]), err)
			return xormdb.RollbackAndLogError(session, "insert into lab_rpki_rtr_full from lab_rpki_rtr_full_log fail: ", err)
		}
	}

	err = UpdateRsyncLogRtrStateEnd(session, labRpkiSyncLogId, state)
	if err != nil {
		belogs.Error("UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd():UpdateRsyncLogRtrStateEnd fail:serialNumber, labRpkiSyncLogId: ",
			serialNumber, labRpkiSyncLogId, err)
		return xormdb.RollbackAndLogError(session, "UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd(): UpdateRsyncLogRtrStateEnd fail: ", err)
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRtr(): CommitSession fail :", err)
		return err
	}
	belogs.Debug("UpdateRtr():CommitSession ok")
	return nil
}

func GetMaxSerialNumber() (serialNumber uint32, err error) {
	sql := `select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1`
	has, err := xormdb.XormEngine.Sql(sql).Get(&serialNumber)
	if err != nil {
		belogs.Error("GetMaxSerialNumber():select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1 fail:", err)
		return serialNumber, err
	}
	if !has {
		// init serialNumber
		serialNumber = 1
	}
	belogs.Debug("GetMaxSerialNumber():select max(sessionserialNumId) lab_rpki_rtr_serial_number, serialNumber :", serialNumber)
	return serialNumber, nil
}
