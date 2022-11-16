package rtrproducer

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"golang.org/x/sync/errgroup"
)

// 1. get all slurm (including had published to rtr)
// 2. get all new roa/asa ( no to rtr)
// 3. start tx: save new roa to db; filter by all slurm; commit tx
// 4. send rtr notify to router
// 5. transfer incr to vc
func rtrUpdateFromSync() (nextStep string, err error) {
	start := time.Now()
	belogs.Info("rtrUpdateFromSync():start")
	var g errgroup.Group
	// update lab_rpki_sync_log set rtring
	labRpkiSyncLogId, err := updateRsyncLogRtrStateStartDb("rtring")
	if err != nil {
		belogs.Error("rtrUpdateFromSync():updateRsyncLogRtrStateStartDb fail:", err, "  time(s):", time.Since(start))
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): labRpkiSyncLogId:", labRpkiSyncLogId, "  time(s):", time.Since(start))

	//get serialNumber
	curSerialNumberModel, newSerialNumberModel, err := getCurAndNewSerialNumberModel()
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getCurAndNewSerialNumberModel fail:", err, "  time(s):", time.Since(start))
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// roa+slurm --> rtrfull/rtrfullog/rtrincr
	g.Go(func() error {
		err1 := rtrUpdateByRoaFromSync(curSerialNumberModel, newSerialNumberModel)
		if err1 != nil {
			belogs.Error("rtrUpdateFromSync():rtrUpdateByRoaFromSync fail:", err, "  time(s):", time.Since(start))
			return err1
		}
		belogs.Info("rtrUpdateFromSync(): rtrUpdateByRoaFromSync pass, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
			"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
		return nil

	})

	// asa --> rtrasafull/rtrasafulllog/rtrasaincr
	g.Go(func() error {
		err1 := rtrUpdateByAsaFromSync(curSerialNumberModel, newSerialNumberModel)
		if err1 != nil {
			belogs.Error("rtrUpdateFromSync(): rtrUpdateByAsaFromSync fail:", err, "  time(s):", time.Since(start))
			return err1
		}
		belogs.Info("rtrUpdateFromSync(): rtrUpdateByAsaFromSync pass, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
			"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

		return nil
	})

	if err := g.Wait(); err != nil {
		belogs.Error("rtrUpdateFromSync(): fail, err:", err, "   time(s):", time.Since(start))
		return "", err
	}

	// update state
	err = updateRsyncLogRtrStateEndDb(labRpkiSyncLogId, "rtred")
	if err != nil {
		belogs.Error("rtrUpdateFromSync():updateRsyncLogRtrStateEndDb fail: newSerialNumber, labRpkiSyncLogId: ",
			jsonutil.MarshalJson(newSerialNumberModel), labRpkiSyncLogId, err, "  time(s):", time.Since(start))
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): updateRsyncLogRtrStateEndDb,  labRpkiSyncLogId:", labRpkiSyncLogId, "  time(s):", time.Since(start))

	// get next step
	nextStep, err = getNextStep()
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getNextStep fail:", err, "  time(s):", time.Since(start))
		return "", err
	}

	belogs.Info("rtrUpdateFromSync():nextStep:", nextStep, " newSerialNumber:", newSerialNumberModel.SerialNumber,
		"  time(s):", time.Since(start))

	belogs.Info("Synchronization and validation processes are completed!!!")
	return nextStep, nil
}

func getCurAndNewSerialNumberModel() (curSerialNumberModel, newSerialNumberModel SerialNumberModel, err error) {
	start := time.Now()
	belogs.Debug("getCurAndNewSerialNumberModel(): ")
	curSerialNumberModel, err = getSerialNumberDb()
	if err != nil {
		belogs.Error("getCurAndNewSerialNumberModel():lab_rpki_rtr_serial_number fail:", err)
		return curSerialNumberModel, newSerialNumberModel, err
	}
	newSerialNumberModel = SerialNumberModel{
		SerialNumber:        curSerialNumberModel.SerialNumber + 1,
		GlobalSerialNumber:  curSerialNumberModel.GlobalSerialNumber + 1,
		SubpartSerialNumber: curSerialNumberModel.SubpartSerialNumber,
	}

	belogs.Info("getCurAndNewSerialNumberModel():  curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
	return curSerialNumberModel, newSerialNumberModel, nil
}

func getNextStep() (nextStep string, err error) {
	myCount, err := getSerialNumberCountDb()
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getSerialNumberCountDb fail:", err)
		return "", err
	}
	if myCount == 0 || myCount == 1 {
		nextStep = "full"
	} else {
		nextStep = "incr"
	}
	return nextStep, nil
}
