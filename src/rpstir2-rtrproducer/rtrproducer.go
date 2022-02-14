package rtrproducer

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

// 1. get all slurm (including had published to rtr)
// 2. get all new roa ( no to rtr)
// 3. start tx: save new roa to db; filter by all slurm; commit tx
// 4. send rtr notify to router
// 5. transfer incr to vc
func rtrUpdateFromSync() (nextStep string, err error) {
	start := time.Now()
	belogs.Info("rtrUpdateFromSync():start")

	// get all slurm
	slurmToRtrFullLogs, err := getAllSlurmsDb()
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getAllSlurmsDb fail:", err)
		return "", err
	}
	belogs.Debug("rtrUpdateFromSync(): slurmToRtrFullLogs:", len(slurmToRtrFullLogs), jsonutil.MarshalJson(slurmToRtrFullLogs))
	belogs.Info("rtrUpdateFromSync(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	//save to lab_rpki_rtr_serial_number, get serialNumber
	curSerialNumberModel, err := getSerialNumberDb()
	if err != nil {
		belogs.Error("rtrUpdateFromSync():lab_rpki_rtr_serial_number fail:", err)
		return "", err
	}
	newSerialNumberModel := SerialNumberModel{
		SerialNumber:        curSerialNumberModel.SerialNumber + 1,
		GlobalSerialNumber:  curSerialNumberModel.GlobalSerialNumber + 1,
		SubpartSerialNumber: curSerialNumberModel.SubpartSerialNumber,
	}

	belogs.Info("rtrUpdateFromSync():  curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Now().Sub(start))

	// get effect slurm from rtrfulllog
	effectSlurmToRtrFullLogs, err := getEffectSlurmsFromSlurm(curSerialNumberModel.SerialNumber, slurmToRtrFullLogs)
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getEffectSlurmsFromSlurm fail:", err)
		return "", err
	}
	belogs.Debug("rtrUpdateFromSync():cur SerialNumber:", curSerialNumberModel.SerialNumber, "  effectSlurmToRtrFullLogs:", len(effectSlurmToRtrFullLogs), jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	belogs.Info("rtrUpdateFromSync():cur SerialNumber:", curSerialNumberModel.SerialNumber, "   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// save chain validate starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := updateRsyncLogRtrStateStartDb("rtring")
	if err != nil {
		belogs.Error("rtrUpdateFromSync():updateRsyncLogRtrStateStartDb fail:", err)
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): labRpkiSyncLogId:", labRpkiSyncLogId)

	// get all roa
	roaToRtrFullLogs, err := getAllRoasDb()
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getAllRoasDb fail:", err)
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs))

	// update to lab_rpki_rtr_full
	err = updateRtrFullLogFromRoaAndSlurmDb(newSerialNumberModel, roaToRtrFullLogs, effectSlurmToRtrFullLogs)
	//err = UpdateRtrDb(slurmDbs)
	if err != nil {
		belogs.Error("rtrUpdateFromSync(): full, updateRtrFullLogFromRoaAndSlurmDb fail:", err)
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): updateRtrFullLogFromRoaAndSlurmDb,  new SerialNumber:", newSerialNumberModel.SerialNumber,
		"  len(roaToRtrFullLogs):", len(roaToRtrFullLogs),
		"  len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), "  time(s):", time.Now().Sub(start))

	// get cur rtrFull
	rtrFullCurs, err := getRtrFullFromRtrFullLogDb(curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getRtrFullFromRtrFullLogDb rtrFullCurs fail: cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): getRtrFullFromRtrFullLogDb len(rtrFullCurs):", len(rtrFullCurs),
		" cur serialNumber:", curSerialNumberModel.SerialNumber, "  time(s):", time.Now().Sub(start))

	// get last rtrFull
	rtrFullNews, err := getRtrFullFromRtrFullLogDb(newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("rtrUpdateFromSync():getRtrFullFromRtrFullLogDb rtrFullNews fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return "", err
	}
	belogs.Info("rtrUpdateFromSync(): getRtrFullFromRtrFullLogDb, len(rtrFullNews):", len(rtrFullNews),
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Now().Sub(start))

	// get rtr incrementals
	rtrIncrementals, err := diffRtrFullToRtrIncremental(rtrFullCurs, rtrFullNews, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("rtrUpdateFromSync():GetRtrFull rtrFullLast fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return "", err
	}
	belogs.Info("rtrUpdateFromSync():diffRtrFullToRtrIncremental, len(rtrIncrementals)", len(rtrIncrementals),
		" new  SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Now().Sub(start))

	// update db
	err = updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb(newSerialNumberModel, rtrIncrementals, labRpkiSyncLogId, "rtred")
	if err != nil {
		belogs.Error("rtrUpdateFromSync():UpdateRtrFullAndIncremental fail:", err)
		return "", err
	}

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
	belogs.Info("rtrUpdateFromSync():updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb, ",
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  len(rtrIncrementals)", len(rtrIncrementals),
		"  time(s):", time.Now().Sub(start))

	belogs.Info("rtrUpdateFromSync(): nextStep:", nextStep, " end time(s):", time.Now().Sub(start).Seconds())
	belogs.Info("Synchronization and validation processes are completed!!!")
	return nextStep, nil
}

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

func getEffectSlurmsFromSlurm(curSerialNumber uint64, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (effectSlurmToRtrFullLogs []model.EffectSlurmToRtrFullLog, err error) {

	belogs.Debug("getEffectSlurmsFromSlurm(): curSerialNumber:", curSerialNumber, " len(slurmToRtrFullLogs): ", len(slurmToRtrFullLogs))
	effectSlurmToRtrFullLogs = make([]model.EffectSlurmToRtrFullLog, 0)
	for i := range slurmToRtrFullLogs {
		var address string
		sourceFrom := model.LabRpkiRtrSourceFrom{
			Source:         "slurm",
			SlurmId:        slurmToRtrFullLogs[i].SlurmId,
			SlurmLogId:     slurmToRtrFullLogs[i].SlurmLogId,
			SlurmLogFileId: slurmToRtrFullLogs[i].SlurmLogFileId,
		}

		if slurmToRtrFullLogs[i].Style == "prefixAssertions" {
			address, _ = iputil.TrimAddressPrefixZero(slurmToRtrFullLogs[i].Address, iputil.GetIpType(slurmToRtrFullLogs[i].Address))
			maxLength := slurmToRtrFullLogs[i].MaxLength
			if maxLength == 0 {
				maxLength = slurmToRtrFullLogs[i].PrefixLength
			}
			effectSlurmToRtrFullLog := model.EffectSlurmToRtrFullLog{
				Id:             slurmToRtrFullLogs[i].Id,
				Style:          slurmToRtrFullLogs[i].Style,
				Asn:            slurmToRtrFullLogs[i].Asn,
				Address:        address,
				PrefixLength:   slurmToRtrFullLogs[i].PrefixLength,
				MaxLength:      maxLength,
				SourceFromJson: jsonutil.MarshalJson(sourceFrom),
			}
			belogs.Debug("getEffectSlurmsFromSlurm():prefixAssertions, slurmToRtrFullLogs[i]:", jsonutil.MarshalJson(slurmToRtrFullLogs),
				"  effectSlurmToRtrFullLog:", jsonutil.MarshalJson(effectSlurmToRtrFullLog))
			effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, effectSlurmToRtrFullLog)

		} else if slurmToRtrFullLogs[i].Style == "prefixFilters" {
			filterSlurms, err := getEffectSlurmsFromSlurmDb(curSerialNumber, slurmToRtrFullLogs[i])
			if err != nil {
				belogs.Error("getEffectSlurmsFromSlurm: diffRtrFullToRtrIncremental fail:",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
				return nil, err
			}
			belogs.Debug("getEffectSlurmsFromSlurm():len(filterSlurms):", len(filterSlurms))
			for filter := range filterSlurms {
				filterSlurms[filter].Style = slurmToRtrFullLogs[i].Style
				filterSlurms[filter].SourceFromJson = jsonutil.MarshalJson(sourceFrom)
			}
			belogs.Debug("getEffectSlurmsFromSlurm():prefixFilters, slurmToRtrFullLogs[i]:", jsonutil.MarshalJson(slurmToRtrFullLogs),
				"  effectSlurmToRtrFullLog:", jsonutil.MarshalJson(filterSlurms))
			effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, filterSlurms...)
		}
	}
	belogs.Debug("getEffectSlurmsFromSlurm():  slurmToRtrFullLogs: ", jsonutil.MarshalJson(slurmToRtrFullLogs),
		"         effectSlurmToRtrFullLogs:", jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	return effectSlurmToRtrFullLogs, nil
}
