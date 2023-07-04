package asa

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
	rtrcommon "rpstir2-rtrproducer/common"
)

func RtrUpdateByAsaFromSync(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Info("RtrUpdateByAsaFromSync():start, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))
	asaToRtrFullLogs, err := getAllAsasDb()
	if err != nil {
		belogs.Error("RtrUpdateByAsaFromSync():getAllAsasDb fail:", err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByAsaFromSync(): len(asaToRtrFullLogs):", len(asaToRtrFullLogs), "  time(s):", time.Since(start))

	// get all slurm
	slurmToRtrFullLogs, err := rtrcommon.GetAllSlurmsDb("asa")
	if err != nil {
		belogs.Error("RtrUpdateByAsaFromSync(): GetAllSlurmsDb fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByAsaFromSync(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Since(start))

	//when both  len are 0, return nil
	if len(asaToRtrFullLogs) == 0 && len(slurmToRtrFullLogs) == 0 {
		belogs.Info("RtrUpdateByAsaFromSync():asa or slurm are both empty")
		return nil
	}

	err = insertRtrAsaFullLogFromAsaDb(newSerialNumberModel.SerialNumber, asaToRtrFullLogs)
	if err != nil {
		belogs.Error("RtrUpdateByAsaFromSync():insertRtrAsaFullLogFromAsaDb fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByAsaFromSync():insertRtrAsaFullLogFromAsaDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(asaToRtrFullLogs):", len(asaToRtrFullLogs), "  time(s):", time.Since(start))

	_, err = rtrcommon.UpdateRtrAsaFullOrFullLogFromSlurmDb("lab_rpki_rtr_asa_full_log", newSerialNumberModel.SerialNumber, slurmToRtrFullLogs, false)
	if err != nil {
		belogs.Error("RtrUpdateByAsaFromSync(): UpdateRtrAsaFullOrFullLogFromSlurmDb lab_rpki_rtr_asa_full_log, fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByAsaFromSync():UpdateRtrAsaFullOrFullLogFromSlurmDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"  len(asaToRtrFullLogs):", len(asaToRtrFullLogs),
		"  len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Since(start))

	// get incrementals from curRtrFullLog and newRtrFullLog different
	rtrAsaIncrementals, err := getRtrAsaIncrementals(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("RtrUpdateByAsaFromSync():getRtrAsaIncrementals fail: curSerialNumberModel:", curSerialNumberModel,
			"   newSerialNumber:", newSerialNumberModel, err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByAsaFromSync():getRtrAsaIncrementals, len(rtrAsaIncrementals)", len(rtrAsaIncrementals),
		"  curSerialNumberModel:", curSerialNumberModel, "   newSerialNumber:", newSerialNumberModel, "  time(s):", time.Since(start))

	err = updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb(newSerialNumberModel, rtrAsaIncrementals)
	if err != nil {
		belogs.Error("RtrUpdateByAsaFromSync():updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb: fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), "   len(rtrAsaIncrementals):", len(rtrAsaIncrementals), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByAsaFromSync(): updateSerialNumberAndRtrAsaFullAndRtrAsaIncrementalDb,  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "  time(s):", time.Since(start))
	return nil
}

func getRtrAsaIncrementals(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental, err error) {
	start := time.Now()
	belogs.Debug("getRtrAsaIncrementals(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel), "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get cur rtrFull
	rtrAsaFullCurs, err := getRtrAsaFullFromRtrFullLogDb(curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrAsaIncrementals():getRtrAsaFullFromRtrFullLogDb rtrAsaFullCurs fail: cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrAsaIncrementals(): getRtrAsaFullFromRtrFullLogDb len(rtrAsaFullCurs):", len(rtrAsaFullCurs),
		" cur serialNumber:", curSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get latest rtrFull
	rtrAsaFullNews, err := getRtrAsaFullFromRtrFullLogDb(newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrAsaIncrementals():getRtrAsaFullFromRtrFullLogDb rtrAsaFullNews fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrAsaIncrementals(): getRtrAsaFullFromRtrFullLogDb, len(rtrAsaFullNews):", len(rtrAsaFullNews),
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get rtr incrementals
	rtrAsaIncrementals, err = diffRtrAsaFullToRtrAsaIncremental(rtrAsaFullCurs, rtrAsaFullNews, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrAsaIncrementals():GetRtrFull rtrFullLast fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrAsaIncrementals():diffRtrAsaFullToRtrAsaIncremental, len(rtrAsaIncrementals)", len(rtrAsaIncrementals),
		" new  SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))
	return rtrAsaIncrementals, nil
}

func diffRtrAsaFullToRtrAsaIncremental(rtrAsaFullCurs, rtrAsaFullNews map[string]model.LabRpkiRtrAsaFull,
	newSerialNumber uint64) (rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental, err error) {
	belogs.Debug("diffRtrAsaFullToRtrAsaIncremental(): len(rtrAsaFullsCurs):", len(rtrAsaFullCurs),
		"   len(rtrAsaFullNews):", len(rtrAsaFullNews), "   newSerialNumber:", newSerialNumber)

	rtrAsaIncrementals = make([]model.LabRpkiRtrAsaIncremental, 0, len(rtrAsaFullCurs))
	for keyNew, valueNew := range rtrAsaFullNews {
		// new exist in cur, then del in cur
		if _, ok := rtrAsaFullCurs[keyNew]; ok {
			belogs.Debug("diffRtrAsaFullToRtrAsaIncremental(): keyNew found in rtrAsaFullCurs:", keyNew,
				"  will del in rtrAsaFullCurs:", jsonutil.MarshalJson(rtrAsaFullCurs[keyNew]),
				"  and will ignore in rtrAsaFullNews:", jsonutil.MarshalJson(valueNew))
			delete(rtrAsaFullCurs, keyNew)
		} else {
			// new is not exist in cur, then this is announce
			rtrAsaIncremental := model.LabRpkiRtrAsaIncremental{
				Style:         "announce",
				CustomerAsn:   valueNew.CustomerAsn,
				ProviderAsn:   valueNew.ProviderAsn,
				AddressFamily: valueNew.AddressFamily,
				SerialNumber:  uint64(newSerialNumber),
				SourceFrom:    valueNew.SourceFrom,
			}
			belogs.Debug("diffRtrAsaFullToRtrAsaIncremental():keyNew not found in rtrAsaFullCurs, valueNew:", jsonutil.MarshalJson(valueNew),
				"   will set as announce incremental:", jsonutil.MarshalJson(rtrAsaIncremental))
			rtrAsaIncrementals = append(rtrAsaIncrementals, rtrAsaIncremental)
		}
	}
	belogs.Debug("diffRtrAsaFullToRtrAsaIncremental(): after announce, remain will as withdraw len(rtrAsaFullCurs):",
		len(rtrAsaFullCurs))
	// remain in cur, is not show in new, so this is withdraw
	for _, valueCur := range rtrAsaFullCurs {
		rtrAsaIncremental := model.LabRpkiRtrAsaIncremental{
			Style:         "withdraw",
			CustomerAsn:   valueCur.CustomerAsn,
			ProviderAsn:   valueCur.ProviderAsn,
			AddressFamily: valueCur.AddressFamily,
			SerialNumber:  uint64(newSerialNumber),
			SourceFrom:    valueCur.SourceFrom,
		}
		belogs.Debug("diffRtrAsaFullToRtrAsaIncremental(): withdraw incremental:",
			jsonutil.MarshalJson(rtrAsaIncremental))
		rtrAsaIncrementals = append(rtrAsaIncrementals, rtrAsaIncremental)
	}
	belogs.Debug("diffRtrAsaFullToRtrAsaIncremental(): newSerialNumber, len(rtrAsaIncrementals):", newSerialNumber, len(rtrAsaIncrementals))
	return rtrAsaIncrementals, nil
}
