package rtrproducer

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
)

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
				belogs.Error("getEffectSlurmsFromSlurm(): getEffectSlurmsFromSlurmDb fail:",
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
