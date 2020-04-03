package model

import (
	"time"
)

type StatisticDb struct {
	FilePath  string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName  string    `json:"fileName" xorm:"fileName varchar(128)"`
	State     string    `json:"state" xorm:"state json"`
	Origin    string    `json:"origin" xorm:"origin json"`
	SyncTime  time.Time `json:"syncTime" xorm:"syncTime datetime"`
	SyncType  string    `json:"syncType" xorm:"syncType varchar(16)"`
	SyncLogId uint64    `json:"syncLogId" xorm:"syncLogId int(10)"`
}

type RirFileCountDb struct {
	Rir      string `json:"rir" xorm:"rir varchar(16)"`
	Count    uint64 `json:"count" xorm:"count int"`
	State    string `json:"state" xorm:"state  varchar(16)"`
	FileType string `json:"fileType" xorm:"fileType varchar(16)"`
}
type RepoFileCountDb struct {
	Repo     string `json:"repo" xorm:"repo varchar(16)"`
	Count    uint64 `json:"count" xorm:"count int"`
	State    string `json:"state" xorm:"state  varchar(16)"`
	FileType string `json:"fileType" xorm:"fileType varchar(16)"`
}
type RirStatisticModel struct {
	// belong to nic
	Rir string `json:"rir"`
	// cer
	CerFileCount FileCount `json:"cerFileCount"`
	// roa
	RoaFileCount FileCount `json:"roaFileCount"`
	// crl
	CrlFileCount FileCount `json:"crlFileCount"`
	// mft
	MftFileCount FileCount `json:"mftFileCount"`

	RepoStatisticModels []RepoStatisticModel `json:"repos"`

	// sync style: rrdp/rsync
	SyncModel SyncModel `json:"syncModel"`
}

//
type RepoStatisticModel struct {
	// belong to nic
	Rir string `json:"rir"`
	// repo url
	Repo string `json:"repo"`

	// cer
	CerFileCount FileCount `json:"cerFileCount"`

	// roa
	RoaFileCount FileCount `json:"roaFileCount"`

	// crl
	CrlFileCount FileCount `json:"crlFileCount"`

	// mft
	MftFileCount FileCount `json:"mftFileCount"`

	FileStates []FileState `json:"fileStates"`
}

type FileCount struct {
	ValidCount   uint64 `json:"validCount"`
	WarningCount uint64 `json:"warningCount"`
	InvalidCount uint64 `json:"invalidCount"`
}

type FileState struct {
	// belong to nic
	Rir string `json:"rir"  xorm:"rir varchar(16)"`
	// repo url
	Repo string `json:"repo"  xorm:"repo varchar(16)"`
	// rsync://+filePath
	Url      string `json:"url"`
	FilePath string `json:"filePath" xorm:"filePath varchar(256)"`
	FileName string `json:"fileName" xorm:"fileName varchar(128)"`
	FileType string `json:"fileType" xorm:"fileType varchar(128)"`
	//invalid/warning
	State            string   `json:"state" xorm:"state varchar(128)"`
	FailDetails      []string `json:"failDetails"`
	StateFailDetails string   `json:"-" xorm:"stateFailDetails json"`
}

type SyncModel struct {
	SyncLogId      uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncStyle      string    `json:"syncStyle" xorm:"syncStyle varchar(32)"`
	RsyncStartTime time.Time `json:"rsyncStartTime" xorm:"rsyncStartTime datetime"`
	RsyncEndTime   time.Time `json:"rsyncEndTime" xorm:"rsyncEndTime datetime"`
	RrdpStartTime  time.Time `json:"rrdpStartTime" xorm:"rrdpStartTime datetime"`
	RrdpEndTime    time.Time `json:"rrdpEndTime" xorm:"rrdpEndTime datetime"`
}
