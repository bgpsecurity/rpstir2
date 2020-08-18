package model

import (
	"time"

	rrdputil "github.com/cpusoft/goutil/rrdputil"
)

// rrdp channel
type RrdpModelChan struct {
	Url  string `json:"url"`
	Dest string `jsong:"dest"`
}

// parse channel
type ParseModelChan struct {
	FilePathNames []string `json:"filePathNames"`
}

// rrdp and parse end channel, may be end
type RrdpParseEndChan struct {
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
}
