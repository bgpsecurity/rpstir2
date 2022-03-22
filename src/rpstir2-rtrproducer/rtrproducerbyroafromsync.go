package rtrproducer

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func rtrUpdateByRoaFromSync(curSerialNumberModel, newSerialNumberModel SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Info("rtrUpdateByRoaFromSync():start, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get all roa
	roaToRtrFullLogs, err := getAllRoasDb()
	if err != nil {
		belogs.Error("rtrUpdateByRoaFromSync():getAllRoasDb fail:", err, "  time(s):", time.Now().Sub(start))
		return err
	}
	if len(roaToRtrFullLogs) == 0 {
		belogs.Info("rtrUpdateByRoaFromSync():roaToRtrFullLogs is empty, time(s):", time.Now().Sub(start))
		return nil
	}
	belogs.Info("rtrUpdateByRoaFromSync(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// get all slurm
	effectSlurmToRtrFullLogs, err := getEffectSlurmToRtrFullLogs(curSerialNumberModel)
	if err != nil {
		belogs.Error("rtrUpdateByRoaFromSync():getEffectSlurmToRtrFullLogs fail:",
			jsonutil.MarshalJson(curSerialNumberModel), err, "  time(s):", time.Now().Sub(start))
		return err
	}
	belogs.Info("rtrUpdateByRoaFromSync(): getEffectSlurmToRtrFullLogs,  curSerialNumberModel:", curSerialNumberModel,
		"  len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// insert to lab_rpki_rtr_full_log
	err = updateRtrFullLogFromRoaAndSlurmDb(newSerialNumberModel, roaToRtrFullLogs, effectSlurmToRtrFullLogs)
	if err != nil {
		belogs.Error("rtrUpdateByRoaFromSync(): full, updateRtrFullLogFromRoaAndSlurmDb fail:", err, "  time(s):", time.Now().Sub(start))
		return err
	}
	belogs.Info("rtrUpdateByRoaFromSync(): updateRtrFullLogFromRoaAndSlurmDb,  new SerialNumber:", newSerialNumberModel.SerialNumber,
		"  len(roaToRtrFullLogs):", len(roaToRtrFullLogs),
		"  len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// get incrementals from curRtrFullLog and newRtrFullLog different
	rtrIncrementals, err := getRtrIncrementals(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("rtrUpdateByRoaFromSync():getRtrIncrementals fail: curSerialNumberModel:", curSerialNumberModel,
			"   newSerialNumber:", newSerialNumberModel, err, "  time(s):", time.Now().Sub(start))
		return err
	}
	belogs.Info("rtrUpdateByRoaFromSync():diffRtrFullToRtrIncremental, len(rtrIncrementals)", len(rtrIncrementals),
		"  curSerialNumberModel:", curSerialNumberModel, "   newSerialNumber:", newSerialNumberModel, "  time(s):", time.Now().Sub(start))

	// save rtrfull/rtrincr to db
	err = updateSerailNumberAndRtrFullAndRtrIncrementalDb(newSerialNumberModel, rtrIncrementals)
	if err != nil {
		belogs.Error("rtrUpdateByRoaFromSync():updateSerailNumberAndRtrFullAndRtrIncrementalDb: fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), "   len(rtrIncrementals):", len(rtrIncrementals), err, "  time(s):", time.Now().Sub(start))
		return err
	}
	belogs.Info("rtrUpdateByRoaFromSync(): updateSerailNumberAndRtrFullAndRtrIncrementalDb,  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrIncrementals):", len(rtrIncrementals), "  time(s):", time.Now().Sub(start))
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

func getEffectSlurmToRtrFullLogs(curSerialNumberModel SerialNumberModel) (effectSlurmToRtrFullLogs []model.EffectSlurmToRtrFullLog, err error) {
	start := time.Now()
	belogs.Debug("getEffectSlurmToRtrFullLogs(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel))

	// get all slurm
	slurmToRtrFullLogs, err := getAllSlurmsDb()
	if err != nil {
		belogs.Error("getEffectSlurmToRtrFullLogs():getAllSlurmsDb fail:", err)
		return nil, err
	}
	belogs.Debug("getEffectSlurmToRtrFullLogs(): slurmToRtrFullLogs:", len(slurmToRtrFullLogs), jsonutil.MarshalJson(slurmToRtrFullLogs))
	belogs.Info("getEffectSlurmToRtrFullLogs(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// get effect slurm from rtrfulllog
	effectSlurmToRtrFullLogs, err = getEffectSlurmsFromSlurm(curSerialNumberModel.SerialNumber, slurmToRtrFullLogs)
	if err != nil {
		belogs.Error("getEffectSlurmToRtrFullLogs():getEffectSlurmsFromSlurm fail:", err)
		return nil, err
	}
	belogs.Debug("getEffectSlurmToRtrFullLogs():cur SerialNumber:", curSerialNumberModel.SerialNumber,
		"   effectSlurmToRtrFullLogs:", len(effectSlurmToRtrFullLogs), jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	belogs.Info("getEffectSlurmToRtrFullLogs():cur SerialNumber:", curSerialNumberModel.SerialNumber,
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))
	return effectSlurmToRtrFullLogs, nil
}

func getRtrIncrementals(curSerialNumberModel, newSerialNumberModel SerialNumberModel) (rtrIncrementals []model.LabRpkiRtrIncremental, err error) {
	start := time.Now()
	belogs.Debug("getRtrIncrementals(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel), "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get cur rtrFull
	rtrFullCurs, err := getRtrFullFromRtrFullLogDb(curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrIncrementals():getRtrFullFromRtrFullLogDb rtrFullCurs fail: cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrIncrementals(): getRtrFullFromRtrFullLogDb len(rtrFullCurs):", len(rtrFullCurs),
		" cur serialNumber:", curSerialNumberModel.SerialNumber, "  time(s):", time.Now().Sub(start))

	// get last rtrFull
	rtrFullNews, err := getRtrFullFromRtrFullLogDb(newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrIncrementals():getRtrFullFromRtrFullLogDb rtrFullNews fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrIncrementals(): getRtrFullFromRtrFullLogDb, len(rtrFullNews):", len(rtrFullNews),
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Now().Sub(start))

	// get rtr incrementals
	rtrIncrementals, err = diffRtrFullToRtrIncremental(rtrFullCurs, rtrFullNews, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrIncrementals():GetRtrFull rtrFullLast fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrIncrementals():diffRtrFullToRtrIncremental, len(rtrIncrementals)", len(rtrIncrementals),
		" new  SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Now().Sub(start))
	return rtrIncrementals, nil
}
