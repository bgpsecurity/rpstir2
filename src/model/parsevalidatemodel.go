package model

import "time"

type ParseValidateBase64Model struct {
	FileName string `json:"fileName"`
	Base64   string `json:"base64"`
}
type CertIdStateModel struct {
	Id       uint64 `json:"id" xorm:"id int"`
	StateStr string `json:"stateStr" xorm:"stateStr varchar"`
	// nextUpdate, or notAfter
	EndTime time.Time `json:"endTime" xorm:"endTime datetime"`
}

type FileModel struct {
	FilePathName     string `json:"filePathName"`     // will save in db
	TempFilePathName string `json:"tempFilePathName"` // using to parse, tempfile
}
