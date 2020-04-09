package model

import ()

type Results struct {
	CerResult Result `json:"cerResult"`
	CrlResult Result `json:"crlResult"`
	MftResult Result `json:"mftResult"`
	RoaResult Result `json:"roaResult"`
}

type Result struct {
	FileType     string `json:"fileType"  xorm:"fileType varchar(32)"`
	ValidCount   uint64 `json:"validCount"  xorm:"validCount int"`
	WarningCount uint64 `json:"warningCount"  xorm:"warningCount int"`
	InvalidCount uint64 `json:"invalidCount"  xorm:"invalidCount int"`
}
