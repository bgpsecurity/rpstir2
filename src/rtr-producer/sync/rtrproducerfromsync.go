package sync

import (
	"time"

	rtrasa "github.com/bgpsecurity/rpstir2/rtr-producer/asa"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	rtrmoa "github.com/bgpsecurity/rpstir2/rtr-producer/moa"
	rtrroa "github.com/bgpsecurity/rpstir2/rtr-producer/roa"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"golang.org/x/sync/errgroup"
)

// 1. get all slurm (including had published to rtr)
// 2. get all new roa/asa ( no to rtr)
// 3. start tx: save new roa to db; filter by all slurm; commit tx
// 4. send rtr notify to router
// 5. transfer incr to vc
func RtrUpdateFromSync() (nextStep string, err error) {
	start := time.Now()
	belogs.Info("RtrUpdateFromSync():start")
	var g errgroup.Group
	// update lab_rpki_sync_log set rtring
	labRpkiSyncLogId, err := updateSyncLogRtrStateStartDb("rtring")
	if err != nil {
		belogs.Error("RtrUpdateFromSync():updateSyncLogRtrStateStartDb fail:", err, "  time(s):", time.Since(start))
		return "", err
	}
	belogs.Info("RtrUpdateFromSync(): labRpkiSyncLogId:", labRpkiSyncLogId, "  time(s):", time.Since(start))

	//get serialNumber
	curSerialNumberModel, newSerialNumberModel, err := getCurAndNewSerialNumberModel()
	if err != nil {
		belogs.Error("RtrUpdateFromSync():getCurAndNewSerialNumberModel fail:", err, "  time(s):", time.Since(start))
		return "", err
	}
	belogs.Info("RtrUpdateFromSync(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// roa+slurm --> rtrfull/rtrfullog/rtrincr
	g.Go(func() error {
		err1 := rtrroa.RtrUpdateByRoaFromSync(curSerialNumberModel, newSerialNumberModel)
		if err1 != nil {
			belogs.Error("RtrUpdateFromSync():RtrUpdateByRoaFromSync fail:", err1, "  time(s):", time.Since(start))
			return err1
		}
		belogs.Info("RtrUpdateFromSync(): RtrUpdateByRoaFromSync pass, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
			"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
		return nil

	})

	// asa --> rtrasafull/rtrasafulllog/rtrasaincr
	g.Go(func() error {
		err1 := rtrasa.RtrUpdateByAsaFromSync(curSerialNumberModel, newSerialNumberModel)
		if err1 != nil {
			belogs.Error("RtrUpdateFromSync(): RtrUpdateByAsaFromSync fail:", err1, "  time(s):", time.Since(start))
			return err1
		}
		belogs.Info("RtrUpdateFromSync(): RtrUpdateByAsaFromSync pass, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
			"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

		return nil
	})

	// moa --> rtrmoafull/rtrmoafulllog/rtrmoaincr
	g.Go(func() error {
		err1 := rtrmoa.RtrUpdateByMoaFromSync(curSerialNumberModel, newSerialNumberModel)
		if err1 != nil {
			belogs.Error("RtrUpdateFromSync(): RtrUpdateByMoaFromSync fail:", err1, "  time(s):", time.Since(start))
			return err1
		}
		belogs.Info("RtrUpdateFromSync(): RtrUpdateByMoaFromSync pass, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
			"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

		return nil
	})

	if err := g.Wait(); err != nil {
		belogs.Error("RtrUpdateFromSync(): fail, err:", err, "   time(s):", time.Since(start))
		return "", err
	}

	// update state
	err = updateSyncLogRtrStateEndDb(labRpkiSyncLogId, "rtred")
	if err != nil {
		belogs.Error("RtrUpdateFromSync():updateSyncLogRtrStateEndDb fail: newSerialNumber, labRpkiSyncLogId: ",
			jsonutil.MarshalJson(newSerialNumberModel), labRpkiSyncLogId, err, "  time(s):", time.Since(start))
		return "", err
	}
	belogs.Info("RtrUpdateFromSync(): updateSyncLogRtrStateEndDb,  labRpkiSyncLogId:", labRpkiSyncLogId, "  time(s):", time.Since(start))

	// get next step
	nextStep, err = getNextStep()
	if err != nil {
		belogs.Error("RtrUpdateFromSync():getNextStep fail:", err, "  time(s):", time.Since(start))
		return "", err
	}

	belogs.Info("RtrUpdateFromSync():nextStep:", nextStep, " newSerialNumber:", newSerialNumberModel.SerialNumber,
		"  time(s):", time.Since(start))

	belogs.Info("Synchronization and validation processes are completed!!!")
	return nextStep, nil
}

func getCurAndNewSerialNumberModel() (curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel, err error) {
	start := time.Now()
	belogs.Debug("getCurAndNewSerialNumberModel(): ")
	curSerialNumberModel, err = rtrcommon.GetSerialNumberDb()
	if err != nil {
		belogs.Error("getCurAndNewSerialNumberModel(): GetSerialNumberDb fail:", err)
		return curSerialNumberModel, newSerialNumberModel, err
	}
	newSerialNumberModel = &rtrcommon.SerialNumberModel{
		SerialNumber:        curSerialNumberModel.SerialNumber + 1,
		GlobalSerialNumber:  curSerialNumberModel.GlobalSerialNumber + 1,
		SubpartSerialNumber: curSerialNumberModel.SubpartSerialNumber,
	}

	belogs.Info("getCurAndNewSerialNumberModel():  curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
	return curSerialNumberModel, newSerialNumberModel, nil
}

func getNextStep() (nextStep string, err error) {
	myCount, err := rtrcommon.GetSerialNumberCountDb()
	if err != nil {
		belogs.Error("RtrUpdateFromSync():GetSerialNumberCountDb fail:", err)
		return "", err
	}
	if myCount == 0 || myCount == 1 {
		nextStep = "full"
	} else {
		nextStep = "incr"
	}
	return nextStep, nil
}
