package rtrserver

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func getSessionIdDb() (sessionId uint16, err error) {
	// lab_rpki_rtr_session, get sessionId
	sql := `select max(sessionId) as sessionId from lab_rpki_rtr_session `
	has, err := xormdb.XormEngine.SQL(sql).Get(&sessionId)
	if err != nil {
		belogs.Error("getSessionIdDb():select max(sessionId) lab_rpki_rtr_session fail:", err)
		return sessionId, err
	}
	if !has {
		belogs.Error("getSessionIdDb():select max(sessionId) lab_rpki_rtr_session have no sessionId:", has)
		return sessionId, errors.New("select max(sessionId) lab_rpki_rtr_session have no sessionId")
	}
	belogs.Debug("getSessionIdDb():select max(sessionId) lab_rpki_rtr_session, sessionId :", sessionId)
	return sessionId, nil
}

func getSpanSerialNumbersDb(clientSerialNumber uint32) (serialNumbers []uint32, err error) {
	serialNumbers = make([]uint32, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_serial_number").Cols("serialNumber").Where("serialNumber > ?", clientSerialNumber).Find(&serialNumbers)
	if err != nil {
		belogs.Error("getSpanSerialNumbersDb():get serialNumbers fail, clientSerialNumber: ", clientSerialNumber, err)
		return serialNumbers, err
	}
	belogs.Debug("getSpanSerialNumbersDb(): clientSerialNumber : ", clientSerialNumber, "   serialNumbers:", serialNumbers)
	return serialNumbers, nil
}

func getRtrXIncrementalDb(clientSerialNumber uint32) (
	rtrIncrementals []model.LabRpkiRtrIncremental,
	rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	rtrMoaIncrementals []model.LabRpkiRtrMoaIncremental,
	sessionId uint16, serialNumber uint32, err error) {

	start := time.Now()
	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).OrderBy("serialNumber").Find(&rtrIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_incremental, clientSerialNumber:", clientSerialNumber,
		"  len(rtrIncrementals):", len(rtrIncrementals))

	rtrAsaIncrementals = make([]model.LabRpkiRtrAsaIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).OrderBy("serialNumber").Find(&rtrAsaIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrAsaIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_asa_incremental, clientSerialNumber:", clientSerialNumber,
		" len(rtrAsaIncrementals):", len(rtrAsaIncrementals))

	rtrMoaIncrementals = make([]model.LabRpkiRtrMoaIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).OrderBy("serialNumber").Find(&rtrMoaIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrMoaIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_asa_incremental, clientSerialNumber:", clientSerialNumber,
		" len(rtrMoaIncrementals):", len(rtrMoaIncrementals))

	sessionId, err = getSessionIdDb()
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():getSessionIdDb fail:", err)
		return nil, nil, nil, sessionId, serialNumber, err
	}

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():getMaxSerialNumberDb fail:", err)
		return nil, nil, nil, sessionId, serialNumber, err
	}

	belogs.Info("getRtrXIncrementalDb():len(rtrIncrementals):", len(rtrIncrementals),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
		"   sessionId:", sessionId, "  serialNumber:", serialNumber,
		"   clientSerialNumber:", clientSerialNumber, "  time(s):", time.Since(start))
	return rtrIncrementals, rtrAsaIncrementals, rtrMoaIncrementals, sessionId, serialNumber, nil
}

func getRtrXFullDb() (rtrFulls []model.LabRpkiRtrFull,
	rtrAsaFulls []model.LabRpkiRtrAsaFull, rtrMoaFulls []model.LabRpkiRtrMoaFull,
	sessionId uint16, serialNumber uint32, err error) {
	start := time.Now()
	/*
		sql := `select id, serialNumber, asn,address, prefixLength,maxLength
		from lab_rpki_rtr_full order by id`
	*/
	rtrFulls = make([]model.LabRpkiRtrFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_full").Cols("id, serialNumber, asn,address, prefixLength,maxLength").
		OrderBy("id").Find(&rtrFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_full fail:", err)
		return nil, nil, nil, 0, 0, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_full, len :", len(rtrFulls))

	rtrAsaFulls = make([]model.LabRpkiRtrAsaFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_asa_full").Cols("id, serialNumber, customerAsn, providerAsns").
		OrderBy("id").Find(&rtrAsaFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_asa_full fail:", err)
		return nil, nil, nil, 0, 0, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_asa_full, len :", len(rtrAsaFulls))

	rtrMoaFulls = make([]model.LabRpkiRtrMoaFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_moa_full").Cols("id, serialNumber, ipv6MappingPrefix, ipv4Prefixes").
		OrderBy("id").Find(&rtrMoaFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_moa_full fail:", err)
		return nil, nil, nil, 0, 0, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_moa_full, len :", len(rtrMoaFulls))

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("getRtrXFullDb():getMaxSerialNumberDb fail:", err)
		return nil, nil, nil, 0, 0, err
	}

	sessionId, err = getSessionIdDb()
	if err != nil {
		belogs.Error("getRtrXFullDb():getSessionIdDb fail:", err)
		return nil, nil, nil, 0, 0, err
	}
	belogs.Info("getRtrXFullDb():len(rtrFulls):", len(rtrFulls),
		"   len(rtrAsaFulls):", len(rtrAsaFulls), " len(rtrMoaFulls):", len(rtrMoaFulls),
		"   sessionId:", sessionId, "  serialNumber:", serialNumber,
		"   time(s):", time.Since(start))
	return rtrFulls, rtrAsaFulls, rtrMoaFulls, sessionId, serialNumber, nil
}

func getSessionIdAndSerialNumberDb() (sessionId uint16, serialNumber uint32, err error) {

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("getSessionIdAndSerialNumberDb():getMaxSerialNumberDb fail:", err)
		return sessionId, serialNumber, err
	}

	sessionId, err = getSessionIdDb()
	if err != nil {
		belogs.Error("getSessionIdAndSerialNumberDb():getSessionIdDb fail:", err)
		return sessionId, serialNumber, err
	}
	belogs.Debug("getSessionIdAndSerialNumberDb(): serialNumber:", serialNumber, "   sessionId:", sessionId)
	return sessionId, serialNumber, nil
}
