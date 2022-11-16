package rtrserver

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
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

func getRtrIncrementalAndSessionIdAndSerialNumberDb(clientSerialNumber uint32) (
	rtrIncrementals []model.LabRpkiRtrIncremental, rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	sessionId uint16, serialNumber uint32, err error) {

	start := time.Now()
	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).Find(&rtrIncrementals)
	if err != nil {
		belogs.Error("getRtrIncrementalAndSessionIdAndSerialNumberDb():get rtrIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrIncrementalAndSessionIdAndSerialNumberDb():select lab_rpki_rtr_incremental,clientSerialNumber, len(rtrIncrementals) :",
		clientSerialNumber, len(rtrIncrementals))

	rtrAsaIncrementals = make([]model.LabRpkiRtrAsaIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).Find(&rtrAsaIncrementals)
	if err != nil {
		belogs.Error("getRtrIncrementalAndSessionIdAndSerialNumberDb():get rtrAsaIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrIncrementalAndSessionIdAndSerialNumberDb():select lab_rpki_rtr_asa_incremental,clientSerialNumber, len(rtrAsaIncrementals) :",
		clientSerialNumber, len(rtrAsaIncrementals))

	sessionId, err = getSessionIdDb()
	if err != nil {
		return nil, nil, sessionId, serialNumber, err
	}

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		return nil, nil, sessionId, serialNumber, err
	}

	belogs.Info("getRtrIncrementalAndSessionIdAndSerialNumberDb():len(rtrIncrementals) :", len(rtrIncrementals),
		"   sessionId:", sessionId, "  serialNumber:", serialNumber,
		"   clientSerialNumber:", clientSerialNumber, "  time(s):", time.Since(start))
	return rtrIncrementals, rtrAsaIncrementals, sessionId, serialNumber, nil
}

func getRtrFullAndSessionIdAndSerialNumberDb() (rtrFulls []model.LabRpkiRtrFull, rtrAsaFulls []model.LabRpkiRtrAsaFull,
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
		belogs.Error("getRtrFullAndSessionIdAndSerialNumberDb():select  lab_rpki_rtr_full fail:", err)
		return nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrFullAndSessionIdAndSerialNumberDb():select lab_rpki_rtr_full, len :", len(rtrFulls))

	rtrAsaFulls = make([]model.LabRpkiRtrAsaFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_asa_full").Cols("id, serialNumber, addressFamily, customerAsn,providerAsns").
		OrderBy("id").Find(&rtrAsaFulls)
	if err != nil {
		belogs.Error("getRtrFullAndSessionIdAndSerialNumberDb():select  lab_rpki_rtr_asa_full fail:", err)
		return nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrFullAndSessionIdAndSerialNumberDb():select lab_rpki_rtr_asa_full, len :", len(rtrAsaFulls))

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		return nil, nil, sessionId, serialNumber, err
	}

	sessionId, err = getSessionIdDb()
	if err != nil {
		return nil, nil, sessionId, serialNumber, err
	}
	belogs.Info("getRtrFullAndSessionIdAndSerialNumberDb():len(rtrFulls) :", len(rtrFulls), "  len(rtrAsaFulls):", len(rtrAsaFulls),
		"   sessionId:", sessionId, "  serialNumber:", serialNumber,
		"   time(s):", time.Since(start))
	return rtrFulls, rtrAsaFulls, sessionId, serialNumber, nil
}
func getSessionIdAndSerialNumberDb() (sessionId uint16, serialNumber uint32, err error) {

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		return sessionId, serialNumber, err
	}

	sessionId, err = getSessionIdDb()
	if err != nil {
		return sessionId, serialNumber, err
	}
	return sessionId, serialNumber, nil
}
