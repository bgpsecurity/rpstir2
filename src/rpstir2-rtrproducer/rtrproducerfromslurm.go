package rtrproducer

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
)

// 1. get all slurm (including had published to rtr)
// 3. start tx: save new roa to db; filter by all slurm; commit tx
// 4. send rtr notify to router
// 5. transfer incr to vc
func rtrUpdateFromSlurm() (err error) {
	start := time.Now()
	belogs.Info("rtrUpdateFromSlurm():start:")

	// get all slurm
	slurmToRtrFullLogs, err := getAllSlurmsDb()
	if err != nil {
		belogs.Error("rtrUpdateFromSlurm():getAllSlurmsDb fail:", err)
		return err
	}
	belogs.Debug("rtrUpdateFromSlurm(): slurmToRtrFullLogs:", len(slurmToRtrFullLogs), jsonutil.MarshalJson(slurmToRtrFullLogs))
	belogs.Info("rtrUpdateFromSlurm(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	if len(slurmToRtrFullLogs) == 0 {
		belogs.Info("rtrUpdateFromSlurm(): len(slurmToRtrFullLogs) is empty, will return 'end' ")
		return nil
	}

	rushNodeIsTopResult := RushNodeIsTopResult{}
	err = httpclient.PostAndUnmarshalResponseModel("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::transferHttpsPort")+
		"/rushnode/istop", "", false, &rushNodeIsTopResult)
	belogs.Info("rtrUpdateFromSlurm(): rushNodeIsTopResult:", rushNodeIsTopResult)
	if err != nil {
		belogs.Error("rtrUpdateFromSlurm():rushNodeIsTopResult fail:", err)
		return err
	}

	//save to lab_rpki_rtr_serial_number, get serialNumber
	curSerialNumberModel, err := getSerialNumberDb()
	if err != nil {
		belogs.Error("rtrUpdateFromSlurm():lab_rpki_rtr_serial_number  fail:", err)
		return err
	}
	newSerialNumberModel := SerialNumberModel{}
	if rushNodeIsTopResult.IsTop == "true" {
		newSerialNumberModel.SerialNumber = curSerialNumberModel.SerialNumber + 1
		newSerialNumberModel.GlobalSerialNumber = curSerialNumberModel.GlobalSerialNumber + 1
		newSerialNumberModel.SubpartSerialNumber = curSerialNumberModel.SubpartSerialNumber
	} else {
		newSerialNumberModel.SerialNumber = curSerialNumberModel.SerialNumber + 1
		newSerialNumberModel.GlobalSerialNumber = curSerialNumberModel.GlobalSerialNumber
		newSerialNumberModel.SubpartSerialNumber = curSerialNumberModel.SubpartSerialNumber + 1
	}

	belogs.Info("rtrUpdateFromSlurm(): rushNodeIsTopResult:", jsonutil.MarshalJson(rushNodeIsTopResult),
		"   curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   time(s):", time.Now().Sub(start))

	// get effect slurm from rtrfulllog
	effectSlurmToRtrFullLogs, err := getEffectSlurmsFromSlurm(curSerialNumberModel.SerialNumber, slurmToRtrFullLogs)
	if err != nil {
		belogs.Error("rtrUpdateFromSlurm():getEffectSlurmsFromSlurm fail:", err)
		return err
	}
	belogs.Debug("rtrUpdateFromSlurm():cur SerialNumber:", curSerialNumberModel.SerialNumber,
		"    effectSlurmToRtrFullLogs:", len(effectSlurmToRtrFullLogs), jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	belogs.Info("rtrUpdateFromSlurm(): cur SerialNumber:", curSerialNumberModel.SerialNumber,
		"    len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// update lab_rpki_rtr_full_log, lab_rpki_rtr_full and lab_rpki_rtr_incremental
	err = updateRtrFullAndFullLogAndIncrementalFromSlurmDb(curSerialNumberModel,
		newSerialNumberModel, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("rtrUpdateFromSlurm():updateRtrFullAndFullLogAndIncrementalFromSlurmDb fail:", err)
		return err
	}
	belogs.Info("rtrUpdateFromSlurm(): end, new SerialNumber:", newSerialNumberModel.GlobalSerialNumber,
		"  time(s):", time.Now().Sub(start))
	return nil
}
