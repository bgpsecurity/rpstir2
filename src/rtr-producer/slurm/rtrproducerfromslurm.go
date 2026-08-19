package slurm

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

// 1. get all slurm (including had published to rtr)
// 3. start tx: save new roa to db; filter by all slurm; commit tx
// 4. send rtr notify to router
// 5. transfer incr to vc
func RtrUpdateFromSlurm() (err error) {
	start := time.Now()
	belogs.Info("RtrUpdateFromSlurm():start:")

	// get all slurm
	prefixSlurmToRtrFullLogs, err := rtrcommon.GetAllSlurmsDb("prefix")
	if err != nil {
		belogs.Error("RtrUpdateFromSlurm(): GetAllSlurmsDb prefix fail:", err)
		return err
	}
	belogs.Debug("RtrUpdateFromSlurm(): prefixSlurmToRtrFullLogs:", len(prefixSlurmToRtrFullLogs), jsonutil.MarshalJson(prefixSlurmToRtrFullLogs))
	belogs.Info("RtrUpdateFromSlurm(): len(prefixSlurmToRtrFullLogs):", len(prefixSlurmToRtrFullLogs), "  time(s):", time.Since(start))

	asaSlurmToRtrFullLogs, err := rtrcommon.GetAllSlurmsDb("asa")
	if err != nil {
		belogs.Error("RtrUpdateFromSlurm(): GetAllSlurmsDb asa fail:", err)
		return err
	}
	belogs.Debug("RtrUpdateFromSlurm(): asaSlurmToRtrFullLogs:", len(asaSlurmToRtrFullLogs), jsonutil.MarshalJson(asaSlurmToRtrFullLogs))
	belogs.Info("RtrUpdateFromSlurm(): len(asaSlurmToRtrFullLogs):", len(asaSlurmToRtrFullLogs), "  time(s):", time.Since(start))

	if len(prefixSlurmToRtrFullLogs) == 0 &&
		len(asaSlurmToRtrFullLogs) == 0 {
		belogs.Info("RtrUpdateFromSlurm(): prefixSlurmToRtrFullLogs and asaSlurmToRtrFullLogs" +
			" all are empty, will return 'end' ")
		return nil
	}

	// GetSerialNumberDb, get serialNumber
	curSerialNumberModel, err := rtrcommon.GetSerialNumberDb()
	if err != nil {
		belogs.Error("RtrUpdateFromSlurm(): GetSerialNumberDb fail:", err)
		return err
	}
	newSerialNumberModel := &rtrcommon.SerialNumberModel{}
	newSerialNumberModel.SerialNumber = curSerialNumberModel.SerialNumber + 1
	newSerialNumberModel.GlobalSerialNumber = curSerialNumberModel.GlobalSerialNumber + 1
	newSerialNumberModel.SubpartSerialNumber = curSerialNumberModel.SubpartSerialNumber

	belogs.Info("RtrUpdateFromSlurm(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   time(s):", time.Since(start))

	// update
	err = updateRtrFullAndFullLogAndIncrementalFromSlurm(curSerialNumberModel, newSerialNumberModel,
		prefixSlurmToRtrFullLogs, asaSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("RtrUpdateFromSlurm():updateRtrFullAndFullLogAndIncrementalFromSlurm fail:", err)
		return err
	}
	belogs.Info("RtrUpdateFromSlurm(): end, new SerialNumber:", newSerialNumberModel.GlobalSerialNumber,
		"  time(s):", time.Since(start))
	return nil
}

func updateRtrFullAndFullLogAndIncrementalFromSlurm(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel,
	prefixSlurmToRtrFullLogs, asaSlurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()

	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"  len(prefixSlurmToRtrFullLogs):", len(prefixSlurmToRtrFullLogs),
		"  len(asaSlurmToRtrFullLogs):", len(asaSlurmToRtrFullLogs))

	err = insertNewSerialNumberDb(newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertNewSerialNumberDb fail:", err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertNewSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))

	// prefix
	err = insertRtrFullLogFromCurSerialNumberDb(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm(): insertRtrFullLogFromCurSerialNumberDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertRtrFullLogFromCurSerialNumberDb, time(s):", time.Since(start))

	effectPrefixSlurm, err := rtrcommon.UpdateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full_log", newSerialNumberModel.SerialNumber, prefixSlurmToRtrFullLogs, true)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrFullOrFullLogFromSlurmDb lab_rpki_rtr_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber, "  len(prefixSlurmToRtrFullLogs):", len(prefixSlurmToRtrFullLogs), err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrFullOrFullLogFromSlurmDb, effectPrefixSlurm:", jsonutil.MarshalJson(effectPrefixSlurm), "  len(prefixSlurmToRtrFullLogs):", len(prefixSlurmToRtrFullLogs), ", time(s):", time.Since(start))

	err = updateRtrXFullByNewSerialNumberDb("lab_rpki_rtr_full", newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrXFullByNewSerialNumberDb fail: new serialNumber:",
			newSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrXFullByNewSerialNumberDb, lab_rpki_rtr_full, time(s):", time.Since(start))

	err = delRtrXFullFromSlurmDb("lab_rpki_rtr_full")
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrXFullFromSlurmDb lab_rpki_rtr_full fail:", err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrXFullFromSlurmDb, lab_rpki_rtr_full, time(s):", time.Since(start))

	_, err = rtrcommon.UpdateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full", newSerialNumberModel.SerialNumber, prefixSlurmToRtrFullLogs, false)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrFullOrFullLogFromSlurmDb lab_rpki_rtr_full fail, new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrFullOrFullLogFromSlurmDb, prefixSlurmToRtrFullLogs:", jsonutil.MarshalJson(prefixSlurmToRtrFullLogs), "   time(s):", time.Since(start))

	err = insertRtrIncrementalByEffectSlurmDb(newSerialNumberModel, effectPrefixSlurm)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertRtrIncrementalByEffectSlurmDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"   len(effectPrefixSlurm):", len(effectPrefixSlurm), err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertRtrIncrementalByEffectSlurmDb, effectPrefixSlurm:", jsonutil.MarshalJson(effectPrefixSlurm), "   time(s):", time.Since(start))

	// asa
	err = insertRtrAsaFullLogFromCurSerialNumberDb(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm(): insertRtrAsaFullLogFromCurSerialNumberDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertRtrAsaFullLogFromCurSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))

	err = updateRtrXFullByNewSerialNumberDb("lab_rpki_rtr_asa_full", newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrXFullByNewSerialNumberDb fail: new serialNumber:",
			newSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrXFullByNewSerialNumberDb, lab_rpki_rtr_asa_full newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))

	err = delRtrXFullFromSlurmDb("lab_rpki_rtr_asa_full")
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrXFullFromSlurmDb lab_rpki_rtr_asa_full fail:", err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrXFullFromSlurmDb, lab_rpki_rtr_asa_full, time(s):", time.Since(start))

	// end
	belogs.Info("updateRtrFullAndFullLogAndIncrementalFromSlurm():CommitSession ok,  time(s):", time.Since(start))
	return nil
}
