package rtrproducer

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
)

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
				belogs.Error("updateRtrFullOrFullLogFromSlurmDb():insert into lab_rpki_rtr_full_log from slurm fail:",
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
