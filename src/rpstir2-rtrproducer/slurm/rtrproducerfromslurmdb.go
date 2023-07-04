package slurm

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	rtrcommon "rpstir2-rtrproducer/common"
)

// tableName: lab_rpki_rtr_full/lab_rpki_rtr_asa_full
func delRtrPrefixOrAsaFullFromSlurmDb(tableName string) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	//  should delete last slurm, then insert new slurm
	sql := `delete from ` + tableName + ` where sourceFrom->'$.source' ='slurm' `
	belogs.Debug("delRtrPrefixOrAsaFullFromSlurmDb():`delete tableName source=slurm, sql:", sql)
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("delRtrPrefixOrAsaFullFromSlurmDb(): delete tableName source=slurm fail, tableName:", tableName, err)
		return xormdb.RollbackAndLogError(session, "delRtrPrefixOrAsaFullFromSlurmDb(): delete tableName source=slurm fail", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delRtrPrefixOrAsaFullFromSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "delRtrPrefixOrAsaFullFromSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("delRtrPrefixOrAsaFullFromSlurmDb(): CommitSession ok: time(s):", time.Since(start))
	return nil

}

func insertRtrIncrementalByEffectSlurmDb(newSerialNumberModel *rtrcommon.SerialNumberModel, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	defer session.Close()
	// lab_rpki_rtr_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_incremental
				 (serialNumber,style,asn,address,   prefixLength,maxLength, sourceFrom) values
				 (?,?,?,?,  ?,?,?)`
	belogs.Debug("insertRtrIncrementalByEffectSlurmDb():lab_rpki_rtr_incremental, len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		if slurmToRtrFullLogs[i].Style == "prefixAssertions" {
			style = "announce"
		} else if slurmToRtrFullLogs[i].Style == "prefixFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style, slurmToRtrFullLogs[i].Asn, slurmToRtrFullLogs[i].Address,
			slurmToRtrFullLogs[i].PrefixLength, slurmToRtrFullLogs[i].MaxLength, slurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrIncrementalByEffectSlurmDb insert into lab_rpki_rtr_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrIncrementalByEffectSlurmDb(): CommitSession ok: len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))
	return nil
}

// tableName: lab_rpki_rtr_full/lab_rpki_rtr_asa_full
func updateRtrPrefixOrAsaFullByNewSerialNumberDb(tableName string, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("updateRtrPrefixOrAsaFullByNewSerialNumberDb(): tableName:", tableName, "  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update ` + tableName + ` set serialNumber= ? `
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateRtrPrefixOrAsaFullByNewSerialNumberDb():update tableName set newSerialNumber fail: tableName:", tableName,
			" new serialNumber:", newSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "updateRtrPrefixOrAsaFullByNewSerialNumberDb update lab_rpki_rtr_full set newSerialNumber fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRtrPrefixOrAsaFullByNewSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateRtrPrefixOrAsaFullByNewSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("updateRtrPrefixOrAsaFullByNewSerialNumberDb(): CommitSession ok:  tableName:", tableName, "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   time(s):", time.Since(start))
	return nil
}

func insertRtrFullLogFromCurSerialNumberDb(curSerialNumberModel *rtrcommon.SerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
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
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))
	return nil

}

func insertRtrAsaFullLogFromCurSerialNumberDb(curSerialNumberModel *rtrcommon.SerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	defer session.Close()
	// lab_rpki_rtr_asa_full_log
	// should ignore last slurm, not insert to new rtr_full_log
	// insert into rtr_full_log and use new slurm update
	sql := `insert into lab_rpki_rtr_asa_full_log (serialNumber,customerAsn,providerAsn,addressFamily,sourceFrom) 
		select ` + convert.ToString(newSerialNumberModel.SerialNumber) + `,customerAsn,providerAsn,addressFamily,sourceFrom from lab_rpki_rtr_asa_full_log
		where serialNumber=? and sourceFrom->'$.source' !='slurm' order by id`
	belogs.Debug("insertRtrAsaFullLogFromCurSerialNumberDb():`insert into lab_rpki_rtr_asa_full_log, sql:", sql)
	_, err = session.Exec(sql, curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_asa_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_asa_full_log fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))
	return nil

}

// insert SerialNumber globalSerialNumber and subpartSerialNumber
func insertNewSerialNumberDb(newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	//save to lab_rpki_rtr_serial_number, get serialNumber
	session, err := xormdb.NewSession()
	defer session.Close()
	start := time.Now()
	err = rtrcommon.InsertSerialNumberDb(session, newSerialNumberModel, start)
	if err != nil {
		belogs.Error("insertNewSerialNumberDb():InsertSerialNumberDb fail,newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "insertNewSerialNumberDb():InsertSerialNumberDb fail:", err)
	}
	belogs.Debug("insertNewSerialNumberDb():InsertSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertNewSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertNewSerialNumberDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertNewSerialNumberDb():CommitSession ok, new SerialNumber:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
	return nil
}

func selectSelfNodeDb() (rushNodeModel model.RushNodeModel, has bool, err error) {
	sql := `select id, nodeName, parentNodeId, url, isSelfUrl from lab_rpki_rush_node where isSelfUrl = 'true' `
	has, err = xormdb.XormEngine.SQL(sql).Get(&rushNodeModel)
	if err != nil {
		belogs.Error("selectSelfNodeDb():lab_rpki_rush_node parentNodeId, fail:", err)
		return rushNodeModel, false, err
	}
	belogs.Debug("selectSelfNodeDb():rushNodeModel:", jsonutil.MarshalJson(rushNodeModel), "  has:", has)
	return rushNodeModel, has, nil
}

func insertRtrAsaIncrementalByEffectSlurmDb(newSerialNumberModel *rtrcommon.SerialNumberModel, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrAsaIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	defer session.Close()
	// lab_rpki_rtr_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_asa_incremental
				 (serialNumber,style,customerAsn,ProviderAsn,  AddressFamily,sourceFrom) values
				 (?,?,?,?,  ?,?)`
	belogs.Debug("insertRtrAsaIncrementalByEffectSlurmDb():lab_rpki_rtr_asa_incremental, len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		if slurmToRtrFullLogs[i].Style == "aspaAssertions" {
			style = "announce"
		} else if slurmToRtrFullLogs[i].Style == "aspaFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style, slurmToRtrFullLogs[i].CustomerAsn, slurmToRtrFullLogs[i].ProviderAsn,
			slurmToRtrFullLogs[i].AddressFamily, slurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrAsaIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_asa_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrAsaIncrementalByEffectSlurmDb insert into lab_rpki_rtr_asa_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsaIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrAsaIncrementalByEffectSlurmDb(): CommitSession ok: len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))
	return nil
}
