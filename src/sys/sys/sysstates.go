package sys

import (
	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	db "sys/db"
	sysmodel "sys/model"
)

func DetailStates() (detailStates map[string]interface{}, err error) {
	syncLog, err := db.GetMaxSyncLog()
	if err != nil {
		belogs.Error("DetailStates():GetMaxSyncLog fail:", err)
		return detailStates, err
	}
	detailStates = make(map[string]interface{}, 0)
	detailStates["state"] = syncLog.State
	if len(syncLog.RsyncState) > 0 {
		syncLogRsyncState := model.SyncLogRsyncState{}
		jsonutil.UnmarshalJson(syncLog.RsyncState, &syncLogRsyncState)
		detailStates["rsyncState"] = syncLogRsyncState
	}
	if len(syncLog.RrdpState) > 0 {
		syncLogRrdpState := model.SyncLogRrdpState{}
		jsonutil.UnmarshalJson(syncLog.RrdpState, &syncLogRrdpState)
		detailStates["rrdpState"] = syncLogRrdpState
	}
	if len(syncLog.DiffState) > 0 {
		syncLogDiffState := model.SyncLogDiffState{}
		jsonutil.UnmarshalJson(syncLog.DiffState, &syncLogDiffState)
		detailStates["diffState"] = syncLogDiffState
	}
	if len(syncLog.ParseValidateState) > 0 {
		syncLogParseValidateState := model.SyncLogParseValidateState{}
		jsonutil.UnmarshalJson(syncLog.ParseValidateState, &syncLogParseValidateState)
		detailStates["parseValidateState"] = syncLogParseValidateState
	}
	if len(syncLog.ChainValidateState) > 0 {
		syncLogChainValidateState := model.SyncLogChainValidateState{}
		jsonutil.UnmarshalJson(syncLog.ChainValidateState, &syncLogChainValidateState)
		detailStates["chainValidateState"] = syncLogChainValidateState
	}
	if len(syncLog.RtrState) > 0 {
		syncLogRtrState := model.SyncLogRtrState{}
		jsonutil.UnmarshalJson(syncLog.RtrState, &syncLogRtrState)
		detailStates["rtrState"] = syncLogRtrState
	}
	belogs.Info("DetailStates(): detailStates:", jsonutil.MarshalJson(detailStates))
	return detailStates, nil
}

func SummaryStates() (summaryStates map[string]interface{}, err error) {
	syncLog, err := db.GetMaxSyncLog()
	if err != nil {
		belogs.Error("SummaryStates():GetMaxSyncLog fail:", err)
		return summaryStates, err
	}

	summaryStates = make(map[string]interface{}, 0)
	state := "running"
	if syncLog.State == "rtred" {
		state = "end"
	}
	summaryStates["state"] = state

	startTime := ""
	if len(syncLog.RsyncState) > 0 {
		syncLogRsyncState := model.SyncLogRsyncState{}
		jsonutil.UnmarshalJson(syncLog.RsyncState, &syncLogRsyncState)
		startTime = convert.Time2String(syncLogRsyncState.StartTime)
	}
	if len(syncLog.RrdpState) > 0 {
		syncLogRrdpState := model.SyncLogRrdpState{}
		jsonutil.UnmarshalJson(syncLog.RrdpState, &syncLogRrdpState)
		startTime = convert.Time2String(syncLogRrdpState.StartTime)
	}
	summaryStates["startTime"] = startTime

	endTime := ""
	if len(syncLog.RtrState) > 0 {
		syncLogRtrState := model.SyncLogRtrState{}
		jsonutil.UnmarshalJson(syncLog.RtrState, &syncLogRtrState)
		endTime = convert.Time2String(syncLogRtrState.EndTime)
	}
	summaryStates["endTime"] = endTime
	belogs.Info("SummaryStates(): summaryStates:", jsonutil.MarshalJson(summaryStates))
	return summaryStates, nil
}

func Results() (results sysmodel.Results, err error) {

	return db.Results()
}
