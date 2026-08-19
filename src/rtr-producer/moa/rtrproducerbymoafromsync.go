package moa

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func RtrUpdateByMoaFromSync(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Info("RtrUpdateByMoaFromSync():start, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))
	moaToRtrFullLogs, err := getAllMoasDb()
	if err != nil {
		belogs.Error("RtrUpdateByMoaFromSync():getAllMoasDb fail:", err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByMoaFromSync(): len(moaToRtrFullLogs):", len(moaToRtrFullLogs), "  time(s):", time.Since(start))

	//when both  len are 0, return nil
	if len(moaToRtrFullLogs) == 0 {
		belogs.Info("RtrUpdateByMoaFromSync():moa or slurm are both empty")
		return nil
	}

	err = insertRtrMoaFullLogFromMoaDb(newSerialNumberModel.SerialNumber, moaToRtrFullLogs)
	if err != nil {
		belogs.Error("RtrUpdateByMoaFromSync():insertRtrMoaFullLogFromMoaDb fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByMoaFromSync():insertRtrMoaFullLogFromMoaDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"   len(moaToRtrFullLogs):", len(moaToRtrFullLogs), "  time(s):", time.Since(start))

	// get incrementals from curRtrFullLog and newRtrFullLog different
	rtrMoaIncrementals, err := getRtrMoaIncrementals(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("RtrUpdateByMoaFromSync():getRtrMoaIncrementals fail: curSerialNumberModel:", curSerialNumberModel,
			"   newSerialNumber:", newSerialNumberModel, err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByMoaFromSync():getRtrMoaIncrementals, len(rtrMoaIncrementals)", len(rtrMoaIncrementals),
		"  curSerialNumberModel:", curSerialNumberModel, "   newSerialNumber:", newSerialNumberModel, "  time(s):", time.Since(start))

	err = updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb(newSerialNumberModel, rtrMoaIncrementals)
	if err != nil {
		belogs.Error("RtrUpdateByMoaFromSync():updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb: fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), "   len(rtrMoaIncrementals):", len(rtrMoaIncrementals), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByMoaFromSync(): updateSerialNumberAndRtrMoaFullAndRtrMoaIncrementalDb,  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrMoaIncrementals):", len(rtrMoaIncrementals), "  time(s):", time.Since(start))
	return nil
}

func getRtrMoaIncrementals(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (rtrMoaIncrementals []model.LabRpkiRtrMoaIncremental, err error) {
	start := time.Now()
	belogs.Debug("getRtrMoaIncrementals(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel), "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get cur rtrFull
	rtrMoaFullCurs, err := getRtrMoaFullFromRtrFullLogDb(curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrMoaIncrementals():getRtrMoaFullFromRtrFullLogDb rtrMoaFullCurs fail: cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrMoaIncrementals(): getRtrMoaFullFromRtrFullLogDb len(rtrMoaFullCurs):", len(rtrMoaFullCurs),
		" cur serialNumber:", curSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get latest rtrFull
	rtrMoaFullNews, err := getRtrMoaFullFromRtrFullLogDb(newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrMoaIncrementals():getRtrMoaFullFromRtrFullLogDb rtrMoaFullNews fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrMoaIncrementals(): getRtrMoaFullFromRtrFullLogDb, len(rtrMoaFullNews):", len(rtrMoaFullNews),
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get rtr incrementals
	rtrMoaIncrementals, err = diffRtrMoaFullToRtrMoaIncremental(rtrMoaFullCurs, rtrMoaFullNews, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrMoaIncrementals():GetRtrFull rtrFullLast fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrMoaIncrementals():diffRtrMoaFullToRtrMoaIncremental, len(rtrMoaIncrementals)", len(rtrMoaIncrementals),
		" new  SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))
	return rtrMoaIncrementals, nil
}

func diffRtrMoaFullToRtrMoaIncremental(rtrMoaFullCurs, rtrMoaFullNews map[string]model.LabRpkiRtrMoaFull,
	newSerialNumber uint64) (rtrMoaIncrementals []model.LabRpkiRtrMoaIncremental, err error) {
	belogs.Debug("diffRtrMoaFullToRtrMoaIncremental(): len(rtrMoaFullsCurs):", len(rtrMoaFullCurs),
		"   len(rtrMoaFullNews):", len(rtrMoaFullNews), "   newSerialNumber:", newSerialNumber)

	rtrMoaIncrementals = make([]model.LabRpkiRtrMoaIncremental, 0, len(rtrMoaFullCurs))
	for keyNew, valueNew := range rtrMoaFullNews {
		// new exist in cur, then del in cur
		if _, ok := rtrMoaFullCurs[keyNew]; ok {
			belogs.Debug("diffRtrMoaFullToRtrMoaIncremental(): keyNew found in rtrMoaFullCurs:", keyNew,
				"  will del in rtrMoaFullCurs:", jsonutil.MarshalJson(rtrMoaFullCurs[keyNew]),
				"  and will ignore in rtrMoaFullNews:", jsonutil.MarshalJson(valueNew))
			delete(rtrMoaFullCurs, keyNew)
		} else {
			// new is not exist in cur, then this is announce
			rtrMoaIncremental := model.LabRpkiRtrMoaIncremental{
				Style:             "announce",
				Ipv6MappingPrefix: valueNew.Ipv6MappingPrefix,
				Ipv4Prefixes:      valueNew.Ipv4Prefixes,
				SerialNumber:      uint64(newSerialNumber),
				SourceFrom:        valueNew.SourceFrom,
			}
			belogs.Debug("diffRtrMoaFullToRtrMoaIncremental():keyNew not found in rtrMoaFullCurs, valueNew:", jsonutil.MarshalJson(valueNew),
				"   will set as announce incremental:", jsonutil.MarshalJson(rtrMoaIncremental))
			rtrMoaIncrementals = append(rtrMoaIncrementals, rtrMoaIncremental)
		}
	}
	belogs.Debug("diffRtrMoaFullToRtrMoaIncremental(): after announce, remain will as withdraw len(rtrMoaFullCurs):",
		len(rtrMoaFullCurs))
	// remain in cur, is not show in new, so this is withdraw
	for _, valueCur := range rtrMoaFullCurs {
		rtrMoaIncremental := model.LabRpkiRtrMoaIncremental{
			Style:             "withdraw",
			Ipv6MappingPrefix: valueCur.Ipv6MappingPrefix,
			Ipv4Prefixes:      valueCur.Ipv4Prefixes,
			SerialNumber:      uint64(newSerialNumber),
			SourceFrom:        valueCur.SourceFrom,
		}
		belogs.Debug("diffRtrMoaFullToRtrMoaIncremental(): withdraw incremental:",
			jsonutil.MarshalJson(rtrMoaIncremental))
		rtrMoaIncrementals = append(rtrMoaIncrementals, rtrMoaIncremental)
	}
	belogs.Debug("diffRtrMoaFullToRtrMoaIncremental(): newSerialNumber, len(rtrMoaIncrementals):", newSerialNumber, len(rtrMoaIncrementals))
	return rtrMoaIncrementals, nil
}
