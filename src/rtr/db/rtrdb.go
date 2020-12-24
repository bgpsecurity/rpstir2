package db

import (
	belogs "github.com/astaxie/beego/logs"
	xormdb "github.com/cpusoft/goutil/xormdb"
)

func GetMaxSerialNumber() (serialNumber uint32, err error) {
	sql := `select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1`
	has, err := xormdb.XormEngine.Sql(sql).Get(&serialNumber)
	if err != nil {
		belogs.Error("GetMaxSerialNumber():select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1 fail:", err)
		return serialNumber, err
	}
	if !has {
		// init serialNumber
		serialNumber = 1
	}
	belogs.Debug("GetMaxSerialNumber():select max(sessionserialNumId) lab_rpki_rtr_serial_number, serialNumber :", serialNumber)
	return serialNumber, nil
}
