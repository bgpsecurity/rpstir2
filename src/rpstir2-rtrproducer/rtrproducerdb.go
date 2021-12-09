package rtrproducer

import (
	"errors"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
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
		i.id		 `
	err = xormdb.XormEngine.SQL(sql).Find(&roaToRtrFullLogs)
	if err != nil {
		belogs.Error("getAllRoasDb(): find fail:", err)
		return roaToRtrFullLogs, err
	}
	belogs.Debug("getAllRoasDb(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs))

	return roaToRtrFullLogs, nil
}

func getAllSlurmsDb() (slurmToRtrFullLogs []model.SlurmToRtrFullLog, err error) {
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
	err = xormdb.XormEngine.SQL(sql).Find(&slurmToRtrFullLogs)
	if err != nil {
		belogs.Error("getAllSlurmsDb(): find fail:", err)
		return slurmToRtrFullLogs, err
	}
	belogs.Debug("getAllSlurmsDb(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	return slurmToRtrFullLogs, nil
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
		"   len(roaToRtrFullLogs):", len(roaToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	err = updateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full_log", newSerialNumberModel.SerialNumber, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrFullLogFromRoaAndSlurmDb(): updateRtrFullOrFullLogFromSlurmDb fail:", err)
		return err
	}
	belogs.Info("updateRtrFullLogFromRoaAndSlurmDb():updateRtrFullOrFullLogFromSlurmDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	belogs.Debug("updateRtrFullLogFromRoaAndSlurmDb():CommitSession ok,   time(s)", time.Now().Sub(start).Seconds())
	return nil
}

// insert SerialNumber globalSerialNumber and subpartSerialNumber
func insertNewSerialNumberDb(serialNumberModel SerialNumberModel) (err error) {
	//save to lab_rpki_rtr_serial_number, get serialNumber
	session, err := xormdb.NewSession()
	defer session.Close()

	belogs.Debug("insertNewSerialNumberDb(): will insert serialNumberModel:", jsonutil.MarshalJson(serialNumberModel))
	now := time.Now()
	sql := ` insert into lab_rpki_rtr_serial_number(
		serialNumber,globalSerialNumber,subpartSerialNumber, createTime)
		 values(?,?,?,?)`
	_, err = session.Exec(sql,
		serialNumberModel.SerialNumber, serialNumberModel.GlobalSerialNumber,
		serialNumberModel.SubpartSerialNumber, now)
	if err != nil {
		belogs.Error("insertNewSerialNumberDb():insert into lab_rpki_rtr_serial_number fail:", jsonutil.MarshalJson(serialNumberModel), err)
		return err
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertNewSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertNewSerialNumberDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertNewSerialNumberDb():CommitSession ok, new SerialNumber:", jsonutil.MarshalJson(serialNumberModel))
	return nil
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
	belogs.Info("insertRtrFullLogFromRoaDb(): CommitSession ok, len(roaToRtrFullLogs): ", len(roaToRtrFullLogs), "   time(s):", time.Now().Sub(start))
	return nil
}

// tableName:lab_rpki_rtr_full_log / lab_rpki_rtr_full
func updateRtrFullOrFullLogFromSlurmDb(tableName string, newSerialNumber uint64,
	effectSlurmToRtrFullLogs []model.EffectSlurmToRtrFullLog) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	// insert into rtr_full_log
	sqlInsertSlurm := `insert  ignore into ` + tableName + `
				(serialNumber,asn,address,prefixLength, maxLength,sourceFrom) values
				(?,?,?,  ?,?,?)`
	sqlDeleteSlurm := `delete from ` + tableName + ` where serialNumber = ? and
		asn=? and address=? and prefixLength=? and maxLength = ? `

	belogs.Debug("updateRtrFullOrFullLogFromSlurmDb(): will insert/del lab_rpki_rtr_full_log from slurmToRtrFullLogs,sqlInsertSlurm:", sqlInsertSlurm,
		"    sqlDeleteSlurm:", sqlDeleteSlurm, " newSerialNumber:", newSerialNumber,
		"    len(slurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs))
	for i := range effectSlurmToRtrFullLogs {
		if effectSlurmToRtrFullLogs[i].Style == "prefixAssertions" {
			affected, err := session.Exec(sqlInsertSlurm,
				newSerialNumber, effectSlurmToRtrFullLogs[i].Asn, effectSlurmToRtrFullLogs[i].Address,
				effectSlurmToRtrFullLogs[i].PrefixLength, effectSlurmToRtrFullLogs[i].MaxLength,
				effectSlurmToRtrFullLogs[i].SourceFromJson)

			if err != nil {
				belogs.Error("updateRtrFullLogFromRoaAndSlurmDb():insert into lab_rpki_rtr_full_log from slurm fail:",
					jsonutil.MarshalJson(effectSlurmToRtrFullLogs[i]), affected, err)
				return xormdb.RollbackAndLogError(session, "updateRtrFullOrFullLogFromSlurmDb(): :insert into lab_rpki_rtr_full_log from slurm fail: ", err)
			}
			addRows, _ := affected.RowsAffected()
			belogs.Debug("updateRtrFullOrFullLogFromSlurmDb(): insert lab_rpki_rtr_full_log from slurmToRtrFullLogs, newSerialNumber:",
				newSerialNumber, ",  effectSlurmToRtrFullLogs[i]: ", jsonutil.MarshalJson(effectSlurmToRtrFullLogs[i]), "  insert  affected:", addRows)
		} else if effectSlurmToRtrFullLogs[i].Style == "prefixFilters" {

			affected, err := session.Exec(sqlDeleteSlurm, newSerialNumber, effectSlurmToRtrFullLogs[i].Asn,
				effectSlurmToRtrFullLogs[i].Address, effectSlurmToRtrFullLogs[i].PrefixLength,
				effectSlurmToRtrFullLogs[i].MaxLength)
			if err != nil {
				belogs.Error("updateRtrFullOrFullLogFromSlurmDb():del lab_rpki_rtr_full_log from slurm fail:",
					jsonutil.MarshalJson(effectSlurmToRtrFullLogs[i]), affected, err)
				return xormdb.RollbackAndLogError(session, "updateRtrFullOrFullLogFromSlurmDb(): :del lab_rpki_rtr_full_log from slurm fail: ", err)
			}
			delRows, _ := affected.RowsAffected()
			belogs.Debug("updateRtrFullOrFullLogFromSlurmDb(): delete lab_rpki_rtr_full_log from slurmToRtrFullLogs, newSerialNumber:",
				newSerialNumber, ",  effectSlurmToRtrFullLogs[i]: ", jsonutil.MarshalJson(effectSlurmToRtrFullLogs[i]), " delete  affected:", delRows)
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRtrFullOrFullLogFromSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateRtrFullOrFullLogFromSlurmDb(): CommitSession fail: ", err)
	}
	belogs.Info("updateRtrFullOrFullLogFromSlurmDb():CommitSession ok,  len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs),
		" time(s):", time.Now().Sub(start))
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
	belogs.Info("getRtrFullFromRtrFullLogDb():map LabRpkiRtrFull, serialNumber, len(rtrFulls):", serialNumber, len(rtrFulls), "   time(s):", time.Now().Sub(start))
	return rtrFulls, nil

}

func updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb(newSerialNumberModel SerialNumberModel, rtrIncrementals []model.LabRpkiRtrIncremental,
	labRpkiSyncLogId uint64, state string) (err error) {
	start := time.Now()

	err = updateSerailNumberAndRtrFullAndRtrIncrementalDb(newSerialNumberModel, rtrIncrementals)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb(): updateSerailNumberAndRtrFullAndRtrIncrementalDb fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), err)
		return err
	}

	err = updateRsyncLogRtrStateEndDb(labRpkiSyncLogId, state)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():updateRsyncLogRtrStateEndDb fail: newSerialNumber, labRpkiSyncLogId: ",
			jsonutil.MarshalJson(newSerialNumberModel), labRpkiSyncLogId, err)
		return err
	}

	belogs.Info("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():CommitSession ok, time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func getSerialNumberDb() (serialNumberModel SerialNumberModel, err error) {
	sql := `select serialNumber, globalSerialNumber, subpartSerialNumber from lab_rpki_rtr_serial_number order by id desc limit 1`
	has, err := xormdb.XormEngine.SQL(sql).Get(&serialNumberModel)
	if err != nil {
		belogs.Error("getSerialNumberDb():select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1 fail:", err)
		return serialNumberModel, err
	}
	if !has {
		// init serialNumber
		serialNumberModel.SerialNumber = 1
		serialNumberModel.GlobalSerialNumber = 1
		serialNumberModel.SubpartSerialNumber = 1
	}
	belogs.Debug("getSerialNumberDb():select max(serialNumberModel) lab_rpki_rtr_serial_number, serialNumberModel :", jsonutil.MarshalJson(serialNumberModel))
	return serialNumberModel, nil
}

func updateRtrFullAndFullLogAndIncrementalFromSlurmDb(curSerialNumberModel SerialNumberModel, newSerialNumberModel SerialNumberModel,
	effectSlurmToRtrFullLogs []model.EffectSlurmToRtrFullLog) (err error) {
	start := time.Now()

	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs))

	err = insertNewSerialNumberDb(newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():insertNewSerialNumberDb fail:", err)
		return err
	}

	err = insertRtrFullLogFromCurSerialNumberDb(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb(): insertRtrFullLogFromCurSerialNumberDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return err
	}

	err = updateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full_log", newSerialNumberModel.SerialNumber, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():updateRtrFullOrFullLogFromSlurmDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return err
	}

	err = updateRtrFullByNewSerailNumberDb(newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():updateRtrFullByNewSerailNumberDb fail: new serialNumber:",
			newSerialNumberModel.SerialNumber, err)
		return err
	}

	err = delRtrFullFromSlurmDb()
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():delRtrFullFromSlurmDb fail:", err)
		return err
	}

	err = updateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full", newSerialNumberModel.SerialNumber, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():updateRtrFullOrFullLogFromSlurmDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return err
	}

	err = insertRtrIncrementalByEffectSlurmDb(newSerialNumberModel, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():insertRtrIncrementalByEffectSlurmDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), err)
		return err
	}

	belogs.Info("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():CommitSession ok,  time(s):", time.Now().Sub(start))
	return nil
}

func getEffectSlurmsFromSlurmDb(curSerialNumber uint64, slurmToRtrFullLog model.SlurmToRtrFullLog) (filterSlurms []model.EffectSlurmToRtrFullLog, err error) {
	filterSlurms = make([]model.EffectSlurmToRtrFullLog, 0)
	start := time.Now()
	belogs.Debug("getEffectSlurmsFromSlurmDb():curSerialNumber:", curSerialNumber,
		" 	slurmToRtrFullLogs:", jsonutil.MarshalJson(slurmToRtrFullLog))

	eng := xormdb.XormEngine.Table("lab_rpki_rtr_full_log").Where(" serialNumber= ? ", curSerialNumber)
	change := false
	if slurmToRtrFullLog.Asn.Int64 > 0 {
		change = true
		eng = eng.And(` asn = ? `, slurmToRtrFullLog.Asn.Int64)
	}
	if slurmToRtrFullLog.PrefixLength > 0 {
		change = true
		eng = eng.And(` prefixLength = ? `, slurmToRtrFullLog.PrefixLength)
	}
	if slurmToRtrFullLog.MaxLength > 0 {
		change = true
		eng = eng.And(` maxLength = ? `, slurmToRtrFullLog.MaxLength)
	}
	if len(slurmToRtrFullLog.Address) > 0 {
		change = true
		address, _ := iputil.TrimAddressPrefixZero(slurmToRtrFullLog.Address, iputil.GetIpType(slurmToRtrFullLog.Address))
		eng = eng.And(` address = ? `, address)
	}
	if !change {
		belogs.Error("getEffectSlurmsFromSlurmDb():not found condition from slurm, continue to next, :",
			jsonutil.MarshalJson(slurmToRtrFullLog))
		return filterSlurms, nil
	}
	err = eng.Cols("asn,address,prefixLength,maxLength").Find(&filterSlurms)
	if err != nil {
		belogs.Error("getEffectSlurmsFromSlurmDb(): get lab_rpki_rtr_full_log fail:", jsonutil.MarshalJson(slurmToRtrFullLog), err)
		return nil, err
	}
	belogs.Info("getEffectSlurmsFromSlurmDb():curSerialNumber:", curSerialNumber,
		"     slurmToRtrFullLogs:", jsonutil.MarshalJson(slurmToRtrFullLog),
		"     filterSlurms:", jsonutil.MarshalJson(filterSlurms), "   time(s):", time.Now().Sub(start))
	return filterSlurms, nil

}

func getSerialNumberCountDb() (myCount uint64, err error) {
	start := time.Now()
	sql := `select count(*) as myCount from lab_rpki_rtr_serial_number`
	has, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getSerialNumberCountDb():select count from lab_rpki_rtr_serial_number, fail:", err)
		return 0, err
	}
	if !has {
		belogs.Error("getSerialNumberCountDb():select count from lab_rpki_rtr_serial_number, !has:")
		return 0, errors.New("has no serialNumber")
	}
	belogs.Info("getSerialNumberCountDb(): myCount: ", myCount, "  time(s):", time.Now().Sub(start))
	return myCount, nil
}

func insertRtrFullLogFromCurSerialNumberDb(curSerialNumberModel SerialNumberModel, newSerialNumberModel SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	defer session.Close()
	// lab_rpki_rtr_full_log
	// should ignore last slurm, not insert to new rtr_full_log
	// insert into rtr_full_log and use new slurm update
	sql := `insert into lab_rpki_rtr_full_log (serialNumber,asn,address,prefixLength, maxLength,sourceFrom) 
		select ` + convert.ToString(newSerialNumberModel.SerialNumber) + `,asn,address,prefixLength, maxLength,sourceFrom from lab_rpki_rtr_full_log
		where serialNumber=? and sourceFrom->'$.source' !='slurm' order by id`
	belogs.Debug("insertRtrFullLogFromCurSerialNumberDb():`insert into lab_rpki_rtr_full_log, sql:", sql)
	_, err = session.Exec(sql, curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("insertRtrFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "insertRtrFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_full_log fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrFullLogFromCurSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrFullLogFromCurSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Now().Sub(start))
	return nil

}

func updateRtrFullByNewSerailNumberDb(newSerialNumberModel SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("updateRtrFullByNewSerailNumberDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update lab_rpki_rtr_full set serialNumber= ? `
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateRtrFullByNewSerailNumberDb():update lab_rpki_rtr_full set newSerialNumber fail: new serialNumber:",
			newSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "updateRtrFullByNewSerailNumberDb update lab_rpki_rtr_full set newSerialNumber fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRtrFullByNewSerailNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateRtrFullByNewSerailNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("updateRtrFullByNewSerailNumberDb(): CommitSession ok: newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   time(s):", time.Now().Sub(start))
	return nil
}

func delRtrFullFromSlurmDb() (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	//  should delete last slurm, then insert new slurm
	sql := `delete from lab_rpki_rtr_full where sourceFrom->'$.source' ='slurm' `
	belogs.Debug("delRtrFullFromSlurmDb():`delete lab_rpki_rtr_full source=slurm, sql:", sql)
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("delRtrFullFromSlurmDb(): delete lab_rpki_rtr_full source=slurm:", err)
		return xormdb.RollbackAndLogError(session, "delRtrFullFromSlurmDb(): delete lab_rpki_rtr_full source=slurm, fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delRtrFullFromSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateRtrFullLogFromRoaAndSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("delRtrFullFromSlurmDb(): CommitSession ok: time(s):", time.Now().Sub(start))
	return nil

}

func insertRtrIncrementalByEffectSlurmDb(newSerialNumberModel SerialNumberModel, effectSlurmToRtrFullLogs []model.EffectSlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	defer session.Close()
	// lab_rpki_rtr_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_incremental
				 (serialNumber,style,asn,address,   prefixLength,maxLength, sourceFrom) values
				 (?,?,?,?,  ?,?,?)`
	belogs.Debug("insertRtrIncrementalByEffectSlurmDb():lab_rpki_rtr_incremental, len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs))
	for i := range effectSlurmToRtrFullLogs {
		if effectSlurmToRtrFullLogs[i].Style == "prefixAssertions" {
			style = "announce"
		} else if effectSlurmToRtrFullLogs[i].Style == "prefixFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style, effectSlurmToRtrFullLogs[i].Asn, effectSlurmToRtrFullLogs[i].Address,
			effectSlurmToRtrFullLogs[i].PrefixLength, effectSlurmToRtrFullLogs[i].MaxLength, effectSlurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(effectSlurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrIncrementalByEffectSlurmDb insert into lab_rpki_rtr_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrIncrementalByEffectSlurmDb(): CommitSession ok: len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Now().Sub(start))
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
		belogs.Error("insertNewSerialNumberDb():insert into lab_rpki_rtr_serial_number fail:", jsonutil.MarshalJson(newSerialNumberModel), err)
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
		"   len(rtrIncrementals):", len(rtrIncrementals), "   time(s):", time.Now().Sub(start))
	return nil
}
