package model

import ()

// rsync channel
type RsyncModelChan struct {
	Url  string `json:"url"`
	Dest string `jsong:"dest"`
}

// parse channel
type ParseModelChan struct {
	FilePathName string `json:"filePathName"`
}

// rsync and parse end channel, may be end
type RsyncParseEndChan struct {
}

type RsyncFileHash struct {
	FilePath    string `json:"filePath" xorm:"filePath varchar(512)"`
	FileName    string `json:"fileName" xorm:"fileName varchar(128)"`
	FileHash    string `json:"fileHash" xorm:"fileHash varchar(512)"`
	JsonAll     string `json:"jsonAll" xorm:"jsonAll json"`
	LastJsonAll string `json:"lastJsonAll" xorm:"lastJsonAll json"`
	// cer/roa/mft/crl, no dot
	FileType string `json:"jsonAll" xorm:"fileType  varchar(16)"`
}
