package model

import (
	"time"
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
