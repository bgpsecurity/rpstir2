package db

import (
	"errors"

	belogs "github.com/astaxie/beego/logs"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

func GetSessionId() (sessionId uint16, err error) {
	// lab_rpki_rtr_session, get sessionId
	sql := `select max(sessionId) as sessionId from lab_rpki_rtr_session `
	has, err := xormdb.XormEngine.Sql(sql).Get(&sessionId)
	if err != nil {
		belogs.Error("GetSessionId():select max(sessionId) lab_rpki_rtr_session fail:", err)
		return sessionId, err
	}
	if !has {
		belogs.Error("GetSessionId():select max(sessionId) lab_rpki_rtr_session have no sessionId:", has)
		return sessionId, errors.New("select max(sessionId) lab_rpki_rtr_session have no sessionId")
	}
	belogs.Debug("GetSessionId():select max(sessionId) lab_rpki_rtr_session, sessionId :", sessionId)
	return sessionId, nil
}

func GetSpanSerialNumbers(clientSerialNumber uint32) (serialNumbers []uint32, err error) {
	serialNumbers = make([]uint32, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_serial_number").Cols("serialNumber").Where("serialNumber > ?", clientSerialNumber).Find(&serialNumbers)
	if err != nil {
		belogs.Error("GetSpanSerialNumbers():get serialNumbers fail, clientSerialNumber: ", clientSerialNumber, err)
		return serialNumbers, err
	}
	belogs.Debug("GetSpanSerialNumbers(): clientSerialNumber : ", clientSerialNumber, "   serialNumbers:", serialNumbers)
	return serialNumbers, nil
}

func GetRtrIncrementalAndSessionIdAndSerialNumber(clientSerialNumber uint32) (rtrIncrementals []model.LabRpkiRtrIncremental,
	sessionId uint16, serialNumber uint32, err error) {
	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).Find(&rtrIncrementals)
	if err != nil {
		belogs.Error("GetRtrIncrementalAndSessionIdAndSerialNumber():get rtrIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return rtrIncrementals, sessionId, serialNumber, err
	}
	belogs.Debug("GetRtrIncrementalAndSessionIdAndSerialNumber():select lab_rpki_rtr_incremental,clientSerialNumber, len(rtrIncrementals) :",
		clientSerialNumber, len(rtrIncrementals))

	sessionId, err = GetSessionId()
	if err != nil {
		return rtrIncrementals, sessionId, serialNumber, err
	}

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = GetMaxSerialNumber()
	if err != nil {
		return rtrIncrementals, sessionId, serialNumber, err
	}

	belogs.Debug("GetRtrIncrementalAndSessionIdAndSerialNumber():len(rtrIncrementals), sessionId, serialNumber,clientSerialNumber :",
		len(rtrIncrementals), sessionId, serialNumber, clientSerialNumber)
	return rtrIncrementals, sessionId, serialNumber, nil
}

func GetRtrFullAndSessionIdAndSerialNumber() (rtrFulls []model.LabRpkiRtrFull, sessionId uint16, serialNumber uint32, err error) {

	/*
		sql := `select id, serialNumber, asn,address, prefixLength,maxLength
		from lab_rpki_rtr_full order by id`
	*/
	err = xormdb.XormEngine.Table("lab_rpki_rtr_full").Cols("id, serialNumber, asn,address, prefixLength,maxLength").
		OrderBy("id").Find(&rtrFulls)

	if err != nil {
		belogs.Error("GetRtrFullAndSerialNumAndSessionId():select  lab_rpki_rtr_full fail:", err)
		return rtrFulls, sessionId, serialNumber, err
	}
	belogs.Debug("GetRtrFullAndSerialNumAndSessionId():select lab_rpki_rtr_full, len :", len(rtrFulls))

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = GetMaxSerialNumber()
	if err != nil {
		return rtrFulls, sessionId, serialNumber, err
	}

	sessionId, err = GetSessionId()
	if err != nil {
		return rtrFulls, sessionId, serialNumber, err
	}
	return rtrFulls, sessionId, serialNumber, nil
}
func GetSessionIdAndSerialNumber() (sessionId uint16, serialNumber uint32, err error) {

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = GetMaxSerialNumber()
	if err != nil {
		return sessionId, serialNumber, err
	}

	sessionId, err = GetSessionId()
	if err != nil {
		return sessionId, serialNumber, err
	}
	return sessionId, serialNumber, nil
}
