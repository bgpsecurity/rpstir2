package rtrproducer

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func getAllAsasDb() ([]model.AsaToRtrFullLog, error) {
	// get lastest syncLogFile.Id
	asaToRtrFullLogs := make([]model.AsaToRtrFullLog, 0)
	asaStrToRtrFullLogs := make([]AsaStrToRtrFullLog, 0)
	sql := `select 	a.id as asaId, a.jsonAll->'$.customerAsns' as customerAsns,
					a.syncLogId,a.syncLogFileId from lab_rpki_asa a
		 	order by a.id `
	err := xormdb.XormEngine.SQL(sql).Find(&asaStrToRtrFullLogs)
	if err != nil {
		belogs.Error("getAllAsasDb(): find fail:", err)
		return nil, err
	}
	belogs.Debug("getAllAsasDb(): len(asaStrToRtrFullLogs):", len(asaStrToRtrFullLogs))
	for i := range asaStrToRtrFullLogs {
		customerAsns := make([]model.CustomerAsn, 0)
		err = jsonutil.UnmarshalJson(asaStrToRtrFullLogs[i].CustomerAsns, &customerAsns)
		if err != nil {
			belogs.Error("getAllAsasDb(): UnmarshalJson asaStrToRtrFullLogs[i].CustomerAsns fail:", asaStrToRtrFullLogs[i].CustomerAsns, err)
			return nil, err
		}
		belogs.Debug("getAllAsasDb(): customerAsns:", jsonutil.MarshalJson(customerAsns))
		for j := range customerAsns {
			asaToRtrFullLog := model.AsaToRtrFullLog{
				AsaId:         asaStrToRtrFullLogs[i].AsaId,
				AddressFamily: customerAsns[j].AddressFamily,
				CustomerAsn:   customerAsns[j].CustomerAsn,
				ProviderAsns:  customerAsns[j].ProviderAsns,
				SyncLogId:     asaStrToRtrFullLogs[i].SyncLogId,
				SyncLogFileId: asaStrToRtrFullLogs[i].SyncLogFileId,
			}
			belogs.Debug("getAllAsasDb(): asaToRtrFullLog:", jsonutil.MarshalJson(asaToRtrFullLog))
			asaToRtrFullLogs = append(asaToRtrFullLogs, asaToRtrFullLog)
		}

	}
	belogs.Info("getAllAsasDb(): len(asaToRtrFullLogs):", len(asaToRtrFullLogs))
	return asaToRtrFullLogs, nil
}

func getRtrAsaFullFromRtrFullLogDb(serialNumber uint64) (rtrAsaFulls map[string]model.LabRpkiRtrAsaFull, err error) {
	start := time.Now()
	belogs.Debug("getRtrAsaFullFromRtrFullLogDb():serialNumber:", serialNumber)
	rtrAsaFs := make([]model.LabRpkiRtrAsaFull, 0)
	sql :=
		`select serialNumber,addressFamily,customerAsn,providerAsns, sourceFrom
	    from lab_rpki_rtr_asa_full_log 
	    where serialNumber = ? 
		order by id `
	err = xormdb.XormEngine.SQL(sql, serialNumber).Find(&rtrAsaFs)
	if err != nil {
		belogs.Error("getRtrAsaFullFromRtrFullLogDb(): find fail: serialNumber: ", serialNumber, err)
		return nil, err
	}
	if len(rtrAsaFs) == 0 {
		belogs.Debug("getRtrAsaFullFromRtrFullLogDb(): len(rtrAsaFs)==0: serialNumber", serialNumber)
		return make(map[string]model.LabRpkiRtrAsaFull, 0), nil
	}
	belogs.Debug("getRtrAsaFullFromRtrFullLogDb():model.LabRpkiRtrAsaFull, serialNumber, len(rtrAsaFs) : ", serialNumber, len(rtrAsaFs))

	rtrAsaFulls = make(map[string]model.LabRpkiRtrAsaFull, len(rtrAsaFs)+50)
	for i := range rtrAsaFs {
		key := convert.ToString(rtrAsaFs[i].AddressFamily.ValueOrZero()) + "_" +
			convert.ToString(rtrAsaFs[i].CustomerAsn) + "_" +
			rtrAsaFs[i].ProviderAsns
		rtrAsaFulls[key] = rtrAsaFs[i]
	}
	belogs.Info("getRtrAsaFullFromRtrFullLogDb():map LabRpkiRtrAsaFull, serialNumber, len(rtrAsaFs):",
		serialNumber, len(rtrAsaFs), "   time(s):", time.Now().Sub(start))
	return rtrAsaFulls, nil

}

func updateRtrAsaFullLogFromAsaDb(newSerialNumberModel SerialNumberModel, asaToRtrFullLogs []model.AsaToRtrFullLog) (err error) {
	//when both  len are 0, return nil
	if len(asaToRtrFullLogs) == 0 {
		belogs.Info("updateRtrAsaFullLogFromAsaDb():asa are empty")
		return nil
	}
	start := time.Now()

	err = insertRtrAsaFullLogFromAsaDb(newSerialNumberModel.SerialNumber, asaToRtrFullLogs)
	if err != nil {
		belogs.Error("updateRtrAsaFullLogFromAsaDb():insertRtrAsaFullLogFromAsaDb fail:", err)
		return err
	}
	belogs.Info("updateRtrAsaFullLogFromAsaDb():insertRtrAsaFullLogFromAsaDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(asaToRtrFullLogs):", len(asaToRtrFullLogs), "  time(s):", time.Now().Sub(start))
	return nil
}

func insertRtrAsaFullLogFromAsaDb(newSerialNumber uint64, asaToRtrFullLogs []model.AsaToRtrFullLog) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	// insert asa into rtr_asa_full_log
	sql := `insert  into lab_rpki_rtr_asa_full_log
				(serialNumber,addressFamily,customerAsn,providerAsns,sourceFrom) values
				(?,?,?,?,?)`
	sourceFrom := model.LabRpkiRtrSourceFrom{
		Source: "sync",
	}
	belogs.Debug("insertRtrAsaFullLogFromAsaDb(): will insert lab_rpki_rtr_asa_full_log from asaToRtrFullLogs, len(asaToRtrFullLogs): ", len(asaToRtrFullLogs))
	for i := range asaToRtrFullLogs {
		sourceFrom.SyncLogId = asaToRtrFullLogs[i].SyncLogId
		sourceFrom.SyncLogFileId = asaToRtrFullLogs[i].SyncLogFileId
		sourceFromJson := jsonutil.MarshalJson(sourceFrom)
		providerAsnsStr := jsonutil.MarshalJson(asaToRtrFullLogs[i].ProviderAsns)
		_, err = session.Exec(sql,
			newSerialNumber, asaToRtrFullLogs[i].AddressFamily, asaToRtrFullLogs[i].CustomerAsn,
			xormdb.SqlNullString(providerAsnsStr), sourceFromJson)
		if err != nil {
			belogs.Error("insertRtrAsaFullLogFromAsaDb():insert into lab_rpki_rtr_asa_full_log from asa fail:",
				jsonutil.MarshalJson(asaToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromAsaDb(): insert into lab_rpki_rtr_asa_full_log fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromAsaDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromAsaDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrAsaFullLogFromAsaDb(): CommitSession ok, len(asaToRtrFullLogs): ", len(asaToRtrFullLogs), "   time(s):", time.Now().Sub(start))
	return nil
}

func updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb(newSerialNumberModel SerialNumberModel,
	rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental) (err error) {
	start := time.Now()
	belogs.Debug("updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals))

	session, err := xormdb.NewSession()
	defer session.Close()

	// serialnumber/rtrasafull/rtrasaincr should in one session
	// insert new serial number
	sql := ` insert into lab_rpki_rtr_serial_number(
		serialNumber,globalSerialNumber,subpartSerialNumber, createTime)
		 values(?,?,?,?)`
	_, err = session.Exec(sql,
		newSerialNumberModel.SerialNumber, newSerialNumberModel.GlobalSerialNumber,
		newSerialNumberModel.SubpartSerialNumber, start)
	if err != nil {
		belogs.Error("updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_serial_number fail:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return err
	}

	// delete and insert into lab_rpki_rtr_asa_full
	sql = `delete from lab_rpki_rtr_asa_full`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():delete lab_rpki_rtr_asa_full fail:", err)
		return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb():delete lab_rpki_rtr_asa_full fail: ", err)
	}

	// insert rtr_full from rtr_full_asa_log
	sql = `
	insert ignore into lab_rpki_rtr_asa_full 
		(serialNumber, addressFamily, customerAsn, 
		providerAsns, sourceFrom ) 
	select serialNumber, addressFamily, customerAsn,  
	    providerAsns, sourceFrom 
	from lab_rpki_rtr_asa_full_log where serialNumber=? order by id`
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_full from lab_rpki_rtr_asa_full_log fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_full from lab_rpki_rtr_asa_full_log fail: ", err)
	}

	// insert into lab_rpki_rtr_asa_incremental
	sql = `insert ignore into lab_rpki_rtr_asa_incremental
		(serialNumber,style,addressFamily,customerAsn,  providerAsns,sourceFrom) values
		(?,?,?,?,  ?,?)`
	for i := range rtrAsaIncrementals {
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, rtrAsaIncrementals[i].Style, rtrAsaIncrementals[i].AddressFamily, rtrAsaIncrementals[i].CustomerAsn,
			rtrAsaIncrementals[i].ProviderAsns, rtrAsaIncrementals[i].SourceFrom)
		if err != nil {
			belogs.Error("updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_incremental fail: newSerialNumber:",
				jsonutil.MarshalJson(newSerialNumberModel), jsonutil.MarshalJson(rtrAsaIncrementals[i]), err)
			return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_full from lab_rpki_rtr_asa_full_log fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): CommitSession fail: ", err)
	}

	belogs.Info("updateSerailNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): CommitSession ok: newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "   time(s):", time.Now().Sub(start))
	return nil
}
