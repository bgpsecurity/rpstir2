package model

import (
	"errors"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	. "github.com/cpusoft/goutil/httpserver"
)

// sync(urls) will send to rrdp/rsync
type SyncUrls struct {
	SyncLogId uint64   `json:"syncLogId"`
	RrdpUrls  []string `json:"rrdpUrls"`
	RsyncUrls []string `json:"rsyncUrls"`
}

// sync(rrdp/rsync) result should return to sync
type SyncResult struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`

	OkUrlsLen uint64   `json:"okUrlsLen"`
	OkUrls    []string `json:"okUrls"`

	//rsync failed
	FailUrls         map[string]string `json:"failUrls"`
	FailUrlsTryCount uint64            `json:"failUrlsTryCount"`

	//parse failed
	FailParseValidateCerts map[string]string `json:"failParseValidateCerts"`

	// diff result
	AddFilesLen      uint64 `json:"addFilesLen"`
	DelFilesLen      uint64 `json:"delFilesLen"`
	UpdateFilesLen   uint64 `json:"updateFilesLen"`
	NoChangeFilesLen uint64 `json:"noChangeFilesLen"`
}

type SyncLogSyncState struct {
	SyncStyle string `json:"syncStyle"`

	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`

	RrdpUrls   []string   `json:"rrdpUrls"`
	RrdpResult SyncResult `json:"rrdpResult"`

	RsyncUrls   []string   `json:"rsyncUrls"`
	RsyncResult SyncResult `json:"rsyncResult"`
}

type SyncLogParseValidateState struct {
	StartTime      time.Time `json:"startTime,omitempty"`
	EndTime        time.Time `json:"endTime,omitempty"`
	ParseFailFiles []string  `json:"parseFailFiles,omitempty"`
}
type SyncLogChainValidateState struct {
	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`
}

type SyncLogRtrState struct {
	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`
}

type ServiceState struct {
	//  start time
	StartTime string `json:"startTime"`

	// is running: true/false. whether the whole sync is complete.
	IsRunning string `json:"isRunning"`
	// current state (only public model): idle/sync/parsevalidate/chainvalidate/rtr
	RunningState  string `json:"runningState"`
	curStateMutex *sync.RWMutex
}

func NewServiceState() *ServiceState {
	ss := &ServiceState{}

	ss.StartTime = convert.Time2StringZone(time.Now())

	ss.IsRunning = "false"
	ss.RunningState = "idle"
	ss.curStateMutex = new(sync.RWMutex)

	return ss
}

// state: sync/parsevalidate/chainvalidate/rtr
// only "sync" need isrunning is "false" and runningState is "idle", and will set isruning is "true"
// others will not change isrunning, and runningState will set state
func (ss *ServiceState) EnterState(state string) (s *ServiceState, err error) {
	ss.curStateMutex.Lock()
	defer ss.curStateMutex.Unlock()
	belogs.Info("EnterState():state:", state, "   ss.isRunning :", ss.IsRunning, "  ss.runningState:", ss.RunningState)

	if state == "sync" {
		if ss.IsRunning == "true" || ss.RunningState != "idle" {
			return nil, errors.New("Synchronization cannot start at the same time")
		}
		ss.IsRunning = "true"
	}

	ss.RunningState = state
	return ss, nil
}

// state: sync/parsevalidate/chainvalidate/rtr
// only "rtr/end" will set isrunning is "false"
// others will not change isurnning, and runingState will set "idle".
func (ss *ServiceState) LeaveState(state string) (s *ServiceState, err error) {
	ss.curStateMutex.Lock()
	defer ss.curStateMutex.Unlock()
	belogs.Debug("LeaveState():state:", state, "   ss.isRunning :", ss.IsRunning, "  ss.runningState:", ss.RunningState)

	if state == "rtr" || state == "end" {
		ss.IsRunning = "false"
	}
	ss.RunningState = "idle"
	return ss, nil
}

type ServiceStateRequest struct {
	// enter/leave/get
	Operate string `json:"operate"`
	State   string `json:"state"`
}
type ServiceStateResponse struct {
	HttpResponse
	ServiceState ServiceState `json:"serviceState"`
}
