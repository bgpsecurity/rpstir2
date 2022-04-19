package rtrproducer

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

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
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurmDb():updateRtrFullByNewSerailNumberDb fail: new serialNumber:",
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
		return xormdb.RollbackAndLogError(session, "delRtrFullFromSlurmDb(): CommitSession fail: ", err)
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
