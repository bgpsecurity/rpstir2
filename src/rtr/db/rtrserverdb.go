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

func ExistSerialNumber(clientSerialNumber uint32) (clientSerialNumberId int64, err error) {

	has, err := xormdb.XormEngine.Table("lab_rpki_rtr_serial_number").Cols("id").Where("serialNumber = ?", clientSerialNumber).Get(&clientSerialNumberId)
	if err != nil {
		belogs.Error("ExistSerialNumber():get id fail:", err)
		return -1, err
	}
	belogs.Debug("ExistSerialNumber():has, clientSerialNumberId :", has, clientSerialNumberId)
	if !has {
		return -1, nil
	}

	return clientSerialNumberId, nil

}

func GetSpanSerialNumbers(id int64) (serialNumbers []int64, err error) {
	serialNumbers = make([]int64, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_serial_number").Cols("id").Where("id > ?", id).Find(&serialNumbers)
	if err != nil {
		belogs.Error("GetSpanSerialNumbers():get serialNumbers fail:", id, err)
		return serialNumbers, err
	}
	return serialNumbers, nil
}

func GetRtrIncrementalAndSessionIdAndSerialNumber(clientSerialNumId int64) (rtrIncrementals []model.LabRpkiRtrIncremental,
	sessionId uint16, serialNumber uint32, err error) {
	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0)
	err = xormdb.XormEngine.Where("id > ?", clientSerialNumId).Find(&rtrIncrementals)
	if err != nil {
		belogs.Error("GetRtrIncrementalAndSessionIdAndSerialNumber():get rtrIncrementals fail:  clientSerialNumId is ", clientSerialNumId, err)
		return rtrIncrementals, sessionId, serialNumber, err
	}
	belogs.Debug("GetRtrIncrementalAndSessionIdAndSerialNumber():select lab_rpki_rtr_incremental, len :", len(rtrIncrementals))

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = GetMaxSerialNumber()
	if err != nil {
		return rtrIncrementals, sessionId, serialNumber, err
	}

	sessionId, err = GetSessionId()
	if err != nil {
		return rtrIncrementals, sessionId, serialNumber, err
	}

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
