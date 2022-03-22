package rtrproducer

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func getSerialNumberCountDb() (myCount uint64, err error) {
	start := time.Now()
	sql := `select count(*) as myCount from lab_rpki_rtr_serial_number`
	has, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getSerialNumberCountDb():select count from lab_rpki_rtr_serial_number, fail:", err)
		return 0, err
	}
	if !has {
		belogs.Error("getSerialNumberCountDb():select count from lab_rpki_rtr_serial_number, !has:")
		return 0, errors.New("has no serialNumber")
	}
	belogs.Info("getSerialNumberCountDb(): myCount: ", myCount, "  time(s):", time.Now().Sub(start))
	return myCount, nil
}
