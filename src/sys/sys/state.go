package sys

import (
	"errors"

	"model"
)

var serviceState *model.ServiceState

func init() {
	serviceState = model.NewServiceState()
}

func ServiceState(ssp model.ServiceStateRequest) (*model.ServiceState, error) {
	if ssp.Operate == "enter" {
		return serviceState.EnterState(ssp.State)
	} else if ssp.Operate == "leave" {
		return serviceState.LeaveState(ssp.State)
	} else if ssp.Operate == "get" {
		return serviceState, nil
	}
	return nil, errors.New("param is error")
}

/*
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
	if len(syncLog.SyncState) > 0 {
		syncLogSyncState := model.SyncLogSyncState{}
		jsonutil.UnmarshalJson(syncLog.SyncState, &syncLogSyncState)
		startTime = convert.Time2String(syncLogSyncState.StartTime)
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
*/
