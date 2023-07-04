package slurm

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
	rtrcommon "rpstir2-rtrproducer/common"
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

	if len(prefixSlurmToRtrFullLogs) == 0 && len(asaSlurmToRtrFullLogs) == 0 {
		belogs.Info("RtrUpdateFromSlurm(): prefixSlurmToRtrFullLogs and asaSlurmToRtrFullLogs both are empty, will return 'end' ")
		return nil
	}

	// check is top of rushnode
	rushNodeModel, has, err := selectSelfNodeDb()
	if err != nil {
		belogs.Error("RtrUpdateFromSlurm():rushNodeIsTopResult fail:", err)
		return err
	}
	belogs.Info("RtrUpdateFromSlurm(): rushNodeModel:", jsonutil.MarshalJson(rushNodeModel), " has:", has)
	isTop := "false"
	if !has || rushNodeModel.ParentNodeId.IsZero() || rushNodeModel.ParentNodeId.ValueOrZero() == 0 {
		isTop = "true"
	}

	// GetSerialNumberDb, get serialNumber
	curSerialNumberModel, err := rtrcommon.GetSerialNumberDb()
	if err != nil {
		belogs.Error("RtrUpdateFromSlurm(): GetSerialNumberDb fail:", err)
		return err
	}
	newSerialNumberModel := &rtrcommon.SerialNumberModel{}
	if isTop == "true" {
		newSerialNumberModel.SerialNumber = curSerialNumberModel.SerialNumber + 1
		newSerialNumberModel.GlobalSerialNumber = curSerialNumberModel.GlobalSerialNumber + 1
		newSerialNumberModel.SubpartSerialNumber = curSerialNumberModel.SubpartSerialNumber
	} else {
		newSerialNumberModel.SerialNumber = curSerialNumberModel.SerialNumber + 1
		newSerialNumberModel.GlobalSerialNumber = curSerialNumberModel.GlobalSerialNumber
		newSerialNumberModel.SubpartSerialNumber = curSerialNumberModel.SubpartSerialNumber + 1
	}

	belogs.Info("RtrUpdateFromSlurm(): isTop:", isTop,
		"   curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   time(s):", time.Since(start))

	// update lab_rpki_rtr_full_log, lab_rpki_rtr_full and lab_rpki_rtr_incremental
	err = updateRtrFullAndFullLogAndIncrementalFromSlurm(curSerialNumberModel,
		newSerialNumberModel, prefixSlurmToRtrFullLogs, asaSlurmToRtrFullLogs)
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
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  len(slurmToRtrFullLogs):", len(prefixSlurmToRtrFullLogs))

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

	err = updateRtrPrefixOrAsaFullByNewSerialNumberDb("lab_rpki_rtr_full", newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrPrefixOrAsaFullByNewSerialNumberDb fail: new serialNumber:",
			newSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrPrefixOrAsaFullByNewSerialNumberDb, lab_rpki_rtr_full, time(s):", time.Since(start))

	err = delRtrPrefixOrAsaFullFromSlurmDb("lab_rpki_rtr_full")
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrPrefixOrAsaFullFromSlurmDb lab_rpki_rtr_full fail:", err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrPrefixOrAsaFullFromSlurmDb, lab_rpki_rtr_full, time(s):", time.Since(start))

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

	effectAsaSlurm, err := rtrcommon.UpdateRtrAsaFullOrFullLogFromSlurmDb("lab_rpki_rtr_asa_full_log", newSerialNumberModel.SerialNumber, asaSlurmToRtrFullLogs, true)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrAsaFullOrFullLogFromSlurmDb fail, lab_rpki_rtr_asa_full_log, new SerialNumber:", newSerialNumberModel.SerialNumber, "   len(asaSlurmToRtrFullLogs):", asaSlurmToRtrFullLogs, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrAsaFullOrFullLogFromSlurmDb, lab_rpki_rtr_asa_full_log effectAsaSlurm:", jsonutil.MarshalJson(effectAsaSlurm), "   len(asaSlurmToRtrFullLogs):", asaSlurmToRtrFullLogs, "   time(s):", time.Since(start))

	err = updateRtrPrefixOrAsaFullByNewSerialNumberDb("lab_rpki_rtr_asa_full", newSerialNumberModel)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrPrefixOrAsaFullByNewSerialNumberDb fail: new serialNumber:",
			newSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():updateRtrPrefixOrAsaFullByNewSerialNumberDb, lab_rpki_rtr_asa_full newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))

	err = delRtrPrefixOrAsaFullFromSlurmDb("lab_rpki_rtr_asa_full")
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrPrefixOrAsaFullFromSlurmDb lab_rpki_rtr_asa_full fail:", err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():delRtrPrefixOrAsaFullFromSlurmDb, lab_rpki_rtr_asa_full, time(s):", time.Since(start))

	_, err = rtrcommon.UpdateRtrAsaFullOrFullLogFromSlurmDb("lab_rpki_rtr_asa_full", newSerialNumberModel.SerialNumber, asaSlurmToRtrFullLogs, false)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrAsaFullOrFullLogFromSlurmDb lab_rpki_asa_rtr_full fail, new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():UpdateRtrAsaFullOrFullLogFromSlurmDb, asaSlurmToRtrFullLogs:", jsonutil.MarshalJson(asaSlurmToRtrFullLogs), "  time(s):", time.Since(start))

	err = insertRtrAsaIncrementalByEffectSlurmDb(newSerialNumberModel, effectAsaSlurm)
	if err != nil {
		belogs.Error("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertRtrAsaIncrementalByEffectSlurmDb fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"   len(effectAsaSlurm):", len(effectAsaSlurm), err)
		return err
	}
	belogs.Debug("updateRtrFullAndFullLogAndIncrementalFromSlurm():insertRtrAsaIncrementalByEffectSlurmDb, effectAsaSlurm:", jsonutil.MarshalJson(effectAsaSlurm), "  time(s):", time.Since(start))

	// end
	belogs.Info("updateRtrFullAndFullLogAndIncrementalFromSlurm():CommitSession ok,  time(s):", time.Since(start))
	return nil
}
