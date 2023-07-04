package model

type DistributedRrdpSnapshotModel struct {
	Id          int    `json:"id"  xorm:"id int"`
	NotifyUrl   string `json:"notifyUrl" xorm:"notifyUrl varchar(512)"`
	SnapshotUrl string `json:"snapshotUrl" xorm:"snapshotUrl varchar(512)"`
	MaxSerial   uint64 `json:"maxSerial" xorm:"maxSerial int"`

	Index     uint64 `json:"index"`
	SyncLogId uint64 `json:"syncLogId"`
}

type DistributedRrdpModel struct {
	Id        int    `json:"id"  xorm:"id int"`
	NotifyUrl string `json:"notifyUrl" xorm:"notifyUrl varchar(512)"`
	DeltaUrl  string `json:"deltaUrl" xorm:"deltaUrl varchar(512)"`
	Serial    uint64 `json:"serial" xorm:"serial int"`

	Index     uint64 `json:"index"`
	SyncLogId uint64 `json:"syncLogId"`
}

// result
type DistributedRrdpSnapshotTotalResult struct {
	SyncLogId   uint64 `json:"syncLogId"`
	SnapshotUrl string `json:"uuid"`

	//cer/roa/mft/crl/asa, not dot
	Total uint64 `json:"total"`
}

type DistributedRrdpResult struct {
	SyncLogId uint64 `json:"syncLogId"`
	DeltaUrl  string `json:"deltaUrl"`

	//cer/roa/mft/crl/asa, not dot
	FileType    string      `json:"fileType"`
	CertModel   interface{} `json:"certModel"`
	StateModel  StateModel  `json:"stateModel"`
	OriginModel OriginModel `json:"originModel"`
}
