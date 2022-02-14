package clear

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
)

func clearStart() {

	go clearSyncLogFileDb()

	go clearRtr()

}

func clearRtr() {
	start := time.Now()
	serialNumber, has, err := getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("clearRtr():serialNumber fail:", err)
		return
	} else if !has {
		belogs.Debug("clearRtr():serialNumber fail:", err)
		return
	}
	deleteSerialNumber := serialNumber - 24
	belogs.Info("clearRtr():serialNumber:", serialNumber, "   deleteSerialNumber:", deleteSerialNumber)
	if serialNumber <= 24 {
		belogs.Info("clearRtr():  serialNumber <= 24:", serialNumber, " time(s):", time.Now().Sub(start).Seconds())
		return
	}
	if deleteSerialNumber <= 0 {
		belogs.Info("clearRtr(): deleteSerialNumber <= 0:", deleteSerialNumber, " time(s):", time.Now().Sub(start).Seconds())
		return
	}

	// delete too old from lab_rpki_rtr_incremental
	err = clearRtrFullLogRtrIncremet("lab_rpki_rtr_incremental", deleteSerialNumber)
	if err != nil {
		belogs.Error("clearRtr():clearRtrFullLogRtrIncremet lab_rpki_rtr_incremental fail:deleteSerialNumber:", deleteSerialNumber, err)
		// no return
	}

	// delete too old from lab_rpki_rtr_full_log
	err = clearRtrFullLogRtrIncremet("lab_rpki_rtr_full_log", deleteSerialNumber)
	if err != nil {
		belogs.Error("clearRtr():clearRtrFullLogRtrIncremet lab_rpki_rtr_full_log fail:deleteSerialNumber:", deleteSerialNumber, err)
		// no return
	}
	belogs.Info("clearRtr(): end, time(s):", time.Now().Sub(start).Seconds())
}
