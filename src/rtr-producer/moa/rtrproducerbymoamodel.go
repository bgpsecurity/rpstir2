package moa

type MoaStrToRtrFullLog struct {
	MoaId         uint64 `json:"roaId" xorm:"roaId int"`
	CustomerAsns  string `json:"customerAsns" xorm:"customerAsns varchar"`
	SyncLogId     uint64 `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64 `json:"syncLogFileId" xorm:"syncLogFileId int"`
}
