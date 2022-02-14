package rtrserver

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func getMaxSerialNumberDb() (serialNumber uint32, err error) {
	sql := `select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1`
	has, err := xormdb.XormEngine.SQL(sql).Get(&serialNumber)
	if err != nil {
		belogs.Error("getMaxSerialNumberDb():select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1 fail:", err)
		return serialNumber, err
	}
	if !has {
		// init serialNumber
		serialNumber = 1
	}
	belogs.Debug("getMaxSerialNumberDb():select max(sessionserialNumId) lab_rpki_rtr_serial_number, serialNumber :", serialNumber)
	return serialNumber, nil
}
