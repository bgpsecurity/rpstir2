package roa

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
	rtrcommon "rpstir2-rtrproducer/common"
)

func RtrUpdateByRoaFromSync(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Info("RtrUpdateByRoaFromSync():start, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get all roa
	roaToRtrFullLogs, err := getAllRoasDb()
	if err != nil {
		belogs.Error("RtrUpdateByRoaFromSync():getAllRoasDb fail:", err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByRoaFromSync(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs), "  time(s):", time.Since(start))

	// get all slurm
	slurmToRtrFullLogs, err := rtrcommon.GetAllSlurmsDb("prefix")
	if err != nil {
		belogs.Error("RtrUpdateByRoaFromSync(): GetAllSlurmsDb prefix, fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByRoaFromSync(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Since(start))

	//when both  len are 0, return nil
	if len(roaToRtrFullLogs) == 0 && len(slurmToRtrFullLogs) == 0 {
		belogs.Info("RtrUpdateByRoaFromSync():roa and slurm all are empty")
		return nil
	}

	err = insertRtrFullLogFromRoaDb(newSerialNumberModel.SerialNumber, roaToRtrFullLogs)
	if err != nil {
		belogs.Error("RtrUpdateByRoaFromSync():insertRtrFullLogFromRoaDb fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByRoaFromSync():insertRtrFullLogFromRoaDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(roaToRtrFullLogs):", len(roaToRtrFullLogs), "  time(s):", time.Since(start))

	_, err = rtrcommon.UpdateRtrFullOrFullLogFromSlurmDb("lab_rpki_rtr_full_log", newSerialNumberModel.SerialNumber, slurmToRtrFullLogs, false)
	if err != nil {
		belogs.Error("RtrUpdateByRoaFromSync(): UpdateRtrFullOrFullLogFromSlurmDb lab_rpki_rtr_full_log, fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByRoaFromSync(): UpdateRtrFullOrFullLogFromSlurmDb lab_rpki_rtr_full_log, new serialNumber:", newSerialNumberModel.SerialNumber,
		"  len(roaToRtrFullLogs):", len(roaToRtrFullLogs),
		"  len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Since(start))

	// get incrementals from curRtrFullLog and newRtrFullLog different
	rtrIncrementals, err := getRtrIncrementals(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("RtrUpdateByRoaFromSync():getRtrIncrementals fail: curSerialNumberModel:", curSerialNumberModel,
			"   newSerialNumber:", newSerialNumberModel, err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByRoaFromSync():diffRtrFullToRtrIncremental, len(rtrIncrementals)", len(rtrIncrementals),
		"  curSerialNumberModel:", curSerialNumberModel, "   newSerialNumber:", newSerialNumberModel, "  time(s):", time.Since(start))

	// save rtrfull/rtrincr to db
	err = updateSerialNumberAndRtrFullAndRtrIncrementalDb(newSerialNumberModel, rtrIncrementals)
	if err != nil {
		belogs.Error("RtrUpdateByRoaFromSync():updateSerialNumberAndRtrFullAndRtrIncrementalDb: fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), "   len(rtrIncrementals):", len(rtrIncrementals), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByRoaFromSync(): updateSerialNumberAndRtrFullAndRtrIncrementalDb,  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrIncrementals):", len(rtrIncrementals), "  time(s):", time.Since(start))
	return nil
}

func diffRtrFullToRtrIncremental(rtrFullCurs, rtrFullNews map[string]model.LabRpkiRtrFull,
	newSerialNumber uint64) (rtrIncrementals []model.LabRpkiRtrIncremental, err error) {
	belogs.Debug("diffRtrFullToRtrIncremental(): len(rtrFullsCurs):", len(rtrFullCurs),
		"   len(rtrFullNews):", len(rtrFullNews), "   newSerialNumber:", newSerialNumber)

	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0, len(rtrFullCurs))

	for keyNew, valueNew := range rtrFullNews {
		// new exist in cur, then del in cur
		if _, ok := rtrFullCurs[keyNew]; ok {
			belogs.Debug("diffRtrFullToRtrIncremental(): keyNew found in rtrFullCurs:", keyNew,
				"  will del in rtrFullCurs:", jsonutil.MarshalJson(rtrFullCurs[keyNew]),
				"  and will ignore in rtrFullNews:", jsonutil.MarshalJson(valueNew))
			delete(rtrFullCurs, keyNew)
		} else {
			// new is not exist in cur, then this is announce
			rtrIncremental := model.LabRpkiRtrIncremental{
				Style:        "announce",
				Asn:          valueNew.Asn,
				Address:      valueNew.Address,
				PrefixLength: valueNew.PrefixLength,
				MaxLength:    valueNew.MaxLength,
				SerialNumber: uint64(newSerialNumber),
				SourceFrom:   valueNew.SourceFrom,
			}
			belogs.Debug("diffRtrFullToRtrIncremental():keyNew not found in rtrFullCurs, valueNew:", jsonutil.MarshalJson(valueNew),
				"   will set as announce incremental:", jsonutil.MarshalJson(rtrIncremental))
			rtrIncrementals = append(rtrIncrementals, rtrIncremental)
		}
	}
	belogs.Debug("diffRtrFullToRtrIncremental(): after announce, remain will as withdraw len(rtrFullCurs):",
		len(rtrFullCurs))
	// remain in cur, is not show in new, so this is withdraw
	for _, valueCur := range rtrFullCurs {
		rtrIncremental := model.LabRpkiRtrIncremental{
			Style:        "withdraw",
			Asn:          valueCur.Asn,
			Address:      valueCur.Address,
			PrefixLength: valueCur.PrefixLength,
			MaxLength:    valueCur.MaxLength,
			SerialNumber: uint64(newSerialNumber),
			SourceFrom:   valueCur.SourceFrom,
		}
		belogs.Debug("diffRtrFullToRtrIncremental(): withdraw incremental:",
			jsonutil.MarshalJson(rtrIncremental))
		rtrIncrementals = append(rtrIncrementals, rtrIncremental)
	}
	belogs.Debug("diffRtrFullToRtrIncremental(): newSerialNumber, len(rtrIncrementals):", newSerialNumber, len(rtrIncrementals))
	return rtrIncrementals, nil
}

func getRtrIncrementals(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (rtrIncrementals []model.LabRpkiRtrIncremental, err error) {
	start := time.Now()
	belogs.Debug("getRtrIncrementals(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel), "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get cur rtrFull
	rtrFullCurs, err := getRtrFullFromRtrFullLogDb(curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrIncrementals():getRtrFullFromRtrFullLogDb rtrFullCurs fail: cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrIncrementals(): getRtrFullFromRtrFullLogDb len(rtrFullCurs):", len(rtrFullCurs),
		" cur serialNumber:", curSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get last rtrFull
	rtrFullNews, err := getRtrFullFromRtrFullLogDb(newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrIncrementals():getRtrFullFromRtrFullLogDb rtrFullNews fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrIncrementals(): getRtrFullFromRtrFullLogDb, len(rtrFullNews):", len(rtrFullNews),
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get rtr incrementals
	rtrIncrementals, err = diffRtrFullToRtrIncremental(rtrFullCurs, rtrFullNews, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrIncrementals():GetRtrFull rtrFullLast fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrIncrementals():diffRtrFullToRtrIncremental, len(rtrIncrementals)", len(rtrIncrementals),
		" new  SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))
	return rtrIncrementals, nil
}
