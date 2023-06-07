package asa

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	rtrcommon "rpstir2-rtrproducer/common"
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
			for k := range customerAsns[j].ProviderAsns {
				asaToRtrFullLog := model.AsaToRtrFullLog{
					AsaId:         asaStrToRtrFullLogs[i].AsaId,
					CustomerAsn:   customerAsns[j].CustomerAsn,
					ProviderAsn:   customerAsns[j].ProviderAsns[k].ProviderAsn,
					AddressFamily: customerAsns[j].ProviderAsns[k].AddressFamily,
					SyncLogId:     asaStrToRtrFullLogs[i].SyncLogId,
					SyncLogFileId: asaStrToRtrFullLogs[i].SyncLogFileId,
				}
				belogs.Debug("getAllAsasDb(): asaToRtrFullLog:", jsonutil.MarshalJson(asaToRtrFullLog))
				asaToRtrFullLogs = append(asaToRtrFullLogs, asaToRtrFullLog)
			}
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
		`select serialNumber,customerAsn,providerAsn,addressFamily,sourceFrom 
	    from lab_rpki_rtr_asa_full_log 
	    where serialNumber = ? 
		order by id `
	err = xormdb.XormEngine.SQL(sql, serialNumber).Find(&rtrAsaFs)
	if err != nil {
		belogs.Error("getRtrAsaFullFromRtrFullLogDb(): get lab_rpki_rtr_asa_full_log fail: serialNumber: ", serialNumber, err)
		return nil, err
	}
	if len(rtrAsaFs) == 0 {
		belogs.Debug("getRtrAsaFullFromRtrFullLogDb(): len(rtrAsaFs)==0: serialNumber", serialNumber)
		return make(map[string]model.LabRpkiRtrAsaFull, 0), nil
	}
	belogs.Debug("getRtrAsaFullFromRtrFullLogDb():model.LabRpkiRtrAsaFull, serialNumber, len(rtrAsaFs) : ", serialNumber, len(rtrAsaFs))

	rtrAsaFulls = make(map[string]model.LabRpkiRtrAsaFull, len(rtrAsaFs)+50)
	for i := range rtrAsaFs {
		key := convert.ToString(rtrAsaFs[i].CustomerAsn) + "_" +
			convert.ToString(rtrAsaFs[i].ProviderAsn) + "_" + convert.ToString(rtrAsaFs[i].AddressFamily.ValueOrZero())
		rtrAsaFulls[key] = rtrAsaFs[i]
	}
	belogs.Info("getRtrAsaFullFromRtrFullLogDb():map LabRpkiRtrAsaFull, serialNumber:",
		serialNumber, "  , len(rtrAsaFs):", len(rtrAsaFs), "   time(s):", time.Since(start))
	return rtrAsaFulls, nil

}

func insertRtrAsaFullLogFromAsaDb(newSerialNumber uint64, asaToRtrFullLogs []model.AsaToRtrFullLog) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	// insert asa into rtr_asa_full_log
	sql := `insert  into lab_rpki_rtr_asa_full_log
				(serialNumber,customerAsn,providerAsn,
					addressFamily,sourceFrom) values
				(?,?,?,    ?,?)`
	sourceFrom := model.LabRpkiRtrSourceFrom{
		Source: "sync",
	}
	belogs.Debug("insertRtrAsaFullLogFromAsaDb(): will insert lab_rpki_rtr_asa_full_log from asaToRtrFullLogs, len(asaToRtrFullLogs): ", len(asaToRtrFullLogs))
	for i := range asaToRtrFullLogs {
		sourceFrom.SyncLogId = asaToRtrFullLogs[i].SyncLogId
		sourceFrom.SyncLogFileId = asaToRtrFullLogs[i].SyncLogFileId
		sourceFromJson := jsonutil.MarshalJson(sourceFrom)
		addressFamilyIpv4, addressFamilyIpv6, err := rtrcommon.ConvertAsaAddressFamilyToRtr(asaToRtrFullLogs[i].AddressFamily)
		if err != nil {
			belogs.Error("insertRtrAsaFullLogFromAsaDb():ConvertAsaAddressFamilyToRtr fail:",
				jsonutil.MarshalJson(asaToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromAsaDb(): ConvertAsaAddressFamilyToRtr fail: ", err)
		}
		belogs.Debug("insertRtrAsaFullLogFromAsaDb(): AddressFamily:", asaToRtrFullLogs[i].AddressFamily,
			"   addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
		if addressFamilyIpv4.Valid {
			_, err = session.Exec(sql,
				newSerialNumber, asaToRtrFullLogs[i].CustomerAsn, asaToRtrFullLogs[i].ProviderAsn,
				addressFamilyIpv4, sourceFromJson)
			if err != nil {
				belogs.Error("insertRtrAsaFullLogFromAsaDb():insert into lab_rpki_rtr_asa_full_log ipv4 from asa fail:",
					jsonutil.MarshalJson(asaToRtrFullLogs[i]), err)
				return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromAsaDb(): insert into lab_rpki_rtr_asa_full_log ipv4 fail: ", err)
			}
		}
		if addressFamilyIpv6.Valid {
			_, err = session.Exec(sql,
				newSerialNumber, asaToRtrFullLogs[i].CustomerAsn, asaToRtrFullLogs[i].ProviderAsn,
				addressFamilyIpv6, sourceFromJson)
			if err != nil {
				belogs.Error("insertRtrAsaFullLogFromAsaDb():insert into lab_rpki_rtr_asa_full_log ipv6 from asa fail:",
					jsonutil.MarshalJson(asaToRtrFullLogs[i]), err)
				return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromAsaDb(): insert into lab_rpki_rtr_asa_full_log ipv6 fail: ", err)
			}
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromAsaDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromAsaDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrAsaFullLogFromAsaDb(): CommitSession ok, len(asaToRtrFullLogs): ", len(asaToRtrFullLogs), "   time(s):", time.Since(start))
	return nil
}

func updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb(newSerialNumberModel *rtrcommon.SerialNumberModel,
	rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental) (err error) {
	start := time.Now()
	belogs.Debug("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals))

	session, err := xormdb.NewSession()
	defer session.Close()

	// serialnumber/rtrasafull/rtrasaincr should in one session
	// insert new serial number
	err = rtrcommon.InsertSerialNumberDb(session, newSerialNumberModel, start)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():InsertSerialNumberDb fail,newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():InsertSerialNumberDb fail:", err)
	}
	belogs.Debug("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():InsertSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// delete and insert into lab_rpki_rtr_asa_full
	sql := `delete from lab_rpki_rtr_asa_full`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():delete lab_rpki_rtr_asa_full fail:", err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():delete lab_rpki_rtr_asa_full fail:", err)
	}
	belogs.Debug("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():delete lab_rpki_rtr_asa_full, time(s):", time.Since(start))

	// insert rtr_asa_full from rtr_full_asa_log
	sql = `
	insert ignore into lab_rpki_rtr_asa_full 
		  (serialNumber, customerAsn, providerAsn, 
		   addressFamily,sourceFrom ) 
	select serialNumber, customerAsn, providerAsn, 
	       addressFamily, sourceFrom 
	from lab_rpki_rtr_asa_full_log where serialNumber=? order by id`
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_full from lab_rpki_rtr_asa_full_log fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_full from lab_rpki_rtr_asa_full_log fail: ", err)
	}
	belogs.Debug("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_full from lab_rpki_rtr_asa_full_log , time(s):", time.Since(start))

	// insert into lab_rpki_rtr_asa_incremental
	sql = `insert ignore into lab_rpki_rtr_asa_incremental
		(serialNumber,style,customerAsn,providerAsn,   addressFamily,sourceFrom) values
		(?,?,?,?,  ?,?)`
	for i := range rtrAsaIncrementals {
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, rtrAsaIncrementals[i].Style, rtrAsaIncrementals[i].CustomerAsn, rtrAsaIncrementals[i].ProviderAsn,
			rtrAsaIncrementals[i].AddressFamily, rtrAsaIncrementals[i].SourceFrom)
		if err != nil {
			belogs.Error("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_incremental fail: newSerialNumber:",
				jsonutil.MarshalJson(newSerialNumberModel), jsonutil.MarshalJson(rtrAsaIncrementals[i]), err)
			return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_incremental fail: ", err)
		}
	}
	belogs.Debug("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb():insert into lab_rpki_rtr_asa_incremental, time(s):", time.Since(start))

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): CommitSession fail: ", err)
	}

	belogs.Info("updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb(): CommitSession ok: newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "   time(s):", time.Since(start))
	return nil
}
