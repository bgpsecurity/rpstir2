package model

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

type ExportRoa struct {
	Asn           int    `json:"asn" xorm:"asn int"`
	AddressPrefix string `json:"addressPrefix" xorm:"addressPrefix varchar(512)"`
	MaxLength     int    `json:"maxLength" xorm:"maxLength int"`
	Rir           string `json:"rir" xorm:"rir varchar(32)"`
	Repo          string `json:"repo" xorm:"repo varchar(64)"`
}
