package rrdp

import (
	"time"

	"github.com/cpusoft/goutil/rrdputil"
)

type RrdpByUrlModel struct {
	NotifyUrl string `json:"notifyUrl"`
	DestPath  string `json:"destPath"`
	HasPath   bool   `json:"hasPath"`
	HasLast   bool   `json:"hasLast"`

	LastSessionId string `json:"lastSessionId"`
	LastCurSerial uint64 `json:"lastCurSerial"`
	SyncLogId     uint64 `json:"syncLogId"`
}

// store snapshot and delta some data
type SnapshotDeltaResult struct {
	NotifyUrl    string
	DestPath     string
	RepoHostPath string

	//snapshot/delta
	RrdpType  string
	RrdpTime  time.Time
	SessionId string
	Serial    uint64
	// when snapshot, lasSerial is 0; when delta, lastSerial is the last serial.
	// it means when there are many deltas, there is still just one lastserial
	LastSerial uint64
	RrdpFiles  []rrdputil.RrdpFile

	SnapshotOrDeltaUrl string
}
type DeltaResult struct {
	DeltaModel rrdputil.DeltaModel `json:"deltaModel"`
	ErrMsg     string              `json:"errMsg"`
}
