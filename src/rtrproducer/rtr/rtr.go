package rtr

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	db "rtrproducer/db"
)

// 1. get all slurm (including had published to rtr)
// 2. get all new roa ( no to rtr)
// 3. start tx: save new roa to db; filter by all slurm; commit tx
// 4. send rtr notify to router
// 5. transfer incr to vc
func RtrUpdate() (nextStep string, err error) {
	start := time.Now()
	belogs.Info("RtrUpdate(): start")
	// save chain validate starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := db.UpdateRsyncLogRtrStateStart("rtring")
	if err != nil {
		belogs.Error("RtrUpdate():GetAllRoas fail:", err)
		return "", err
	}

	// get all roa
	roaToRtrFullLogs, err := db.GetAllRoas()
	if err != nil {
		belogs.Error("RtrUpdate():GetAllRoas fail:", err)
		return "", err
	}
	belogs.Info("RtrUpdate(): len(roaToRtrFullLogs):", len(roaToRtrFullLogs))

	// get all rtr
	slurmToRtrFullLogs, err := db.GetAllSlurms()
	if err != nil {
		belogs.Error("RtrUpdate():GetAllSlurms fail:", err)
		return "", err
	}
	belogs.Info("RtrUpdate(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	// update to lab_rpki_rtr_full and lab_rpki_rtr_incremental
	serialNumber, err := db.UpdateRtrFullLog(roaToRtrFullLogs, slurmToRtrFullLogs)
	//err = UpdateRtrDb(slurmDbs)
	if err != nil {
		belogs.Error("RtrUpdate():UpdateRtrFullLog fail:", err)
		return "", err
	}
	belogs.Debug("RtrUpdate(): serialNumber:", serialNumber)

	// get cur rtrFull
	rtrFullCurs, err := db.GetRtrFullFromRtrFullLog(serialNumber)
	if err != nil {
		belogs.Error("RtrUpdate():GetRtrFullFromRtrFullLog rtrFullCurs fail: serialNumber:", serialNumber, err)
		return "", err
	}
	belogs.Info("RtrUpdate(): len(rtrFullCurs), serialNumber:", len(rtrFullCurs), serialNumber)

	// get last rtrFull
	rtrFullLasts, err := db.GetRtrFullFromRtrFullLog(serialNumber - 1)
	if err != nil {
		belogs.Error("RtrUpdate():GetRtrFullFromRtrFullLog rtrFullLasts fail: serialNumber-1:", serialNumber-1, err)
		return "", err
	}
	belogs.Info("RtrUpdate(): len(rtrFullLasts), serialNumber-1:", len(rtrFullLasts), serialNumber-1)

	// get rtr incrementals
	rtrIncrementals, err := diffRtrFullToRtrIncremental(rtrFullCurs, rtrFullLasts, serialNumber)
	if err != nil {
		belogs.Error("RtrUpdate():GetRtrFull rtrFullLast fail: serialNumber-1:", serialNumber-1, err)
		return "", err
	}
	belogs.Info("RtrUpdate(): len(rtrIncrementals), serialNumber:", len(rtrIncrementals), serialNumber)

	// update db
	err = db.UpdateRtrFullAndIncrementalAndRsyncLogRtrStateEnd(serialNumber, rtrIncrementals, labRpkiSyncLogId, "rtred")
	if err != nil {
		belogs.Error("RtrUpdate():UpdateRtrFullAndIncremental fail:", err)
		return "", err
	}

	belogs.Info("RtrUpdate(): end   time(s):", time.Now().Sub(start).Seconds())
	belogs.Info("Synchronization and validation processes are completed!!!")
	return "end", nil
}

func diffRtrFullToRtrIncremental(rtrFullCurs, rtrFullLasts map[string]model.LabRpkiRtrFull,
	serialNumber uint32) (rtrIncrementals []model.LabRpkiRtrIncremental, err error) {
	belogs.Debug("diffRtrFullToRtrIncremental(): len(rtrFullsCurs):", len(rtrFullCurs),
		"   len(rtrFullLasts):", len(rtrFullLasts), "   serialNumber:", serialNumber)

	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0, len(rtrFullCurs))

	// all are add

	for keyCur, valueCur := range rtrFullCurs {
		// cur exist in last, then del in last
		if _, ok := rtrFullLasts[keyCur]; ok {
			delete(rtrFullLasts, keyCur)
		} else {
			// cur is not exist in last ,then this is announce
			rtrIncremental := model.LabRpkiRtrIncremental{
				Style:        "announce",
				Asn:          valueCur.Asn,
				Address:      valueCur.Address,
				PrefixLength: valueCur.PrefixLength,
				MaxLength:    valueCur.MaxLength,
				SerialNumber: uint64(serialNumber),
				SourceFrom:   valueCur.SourceFrom,
			}
			belogs.Debug("diffRtrFullToRtrIncremental(): announce incremental:",
				jsonutil.MarshalJson(rtrIncremental))
			rtrIncrementals = append(rtrIncrementals, rtrIncremental)
		}
	}
	belogs.Debug("diffRtrFullToRtrIncremental(): after announce, remain len(rtrFullLasts) :",
		len(rtrFullLasts))
	// remain in last, is not show in cur, so this is withdraw
	for _, valueLast := range rtrFullLasts {
		rtrIncremental := model.LabRpkiRtrIncremental{
			Style:        "withdraw",
			Asn:          valueLast.Asn,
			Address:      valueLast.Address,
			PrefixLength: valueLast.PrefixLength,
			MaxLength:    valueLast.MaxLength,
			SerialNumber: uint64(serialNumber),
			SourceFrom:   valueLast.SourceFrom,
		}
		belogs.Debug("diffRtrFullToRtrIncremental(): withdraw incremental:",
			jsonutil.MarshalJson(rtrIncremental))
		rtrIncrementals = append(rtrIncrementals, rtrIncremental)
	}
	belogs.Debug("diffRtrFullToRtrIncremental(): serialNumber,len(rtrIncrementals):", serialNumber, len(rtrIncrementals))
	return rtrIncrementals, nil
}
