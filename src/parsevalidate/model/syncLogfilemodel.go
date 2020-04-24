package model

import (
	"model"
)

type SyncLogFileModel struct {
	Id         uint64           `json:"id" xorm:"pk autoincr"`
	SyncLogId  uint64           `json:"syncLogId" xorm:"syncLogId int"`
	FilePath   string           `json:"filePath" xorm:"filePath varchar(512)"`
	FileName   string           `json:"fileName" xorm:"fileName varchar(128)"`
	FileType   string           `json:"fileType" xorm:"fileType varchar(16)"`
	SyncType   string           `json:"syncType" xorm:"syncType varchar(16)"`
	CertModel  interface{}      `json:"-"`
	StateModel model.StateModel `json:"-"`
	JsonAll    string           `json:"jsonAll"`
	//cerId / mftId / roaId / crlId
	CertId uint64 `json:"certId"`
}
type SyncLogFileModels struct {
	SyncLogId                  uint64             `json:"syncLogId"`
	AddCerSyncLogFileModels    []SyncLogFileModel `json:"addCerSyncLogFileModels"`
	UpdateCerSyncLogFileModels []SyncLogFileModel `json:"updateCerSyncLogFileModels"`
	DelCerSyncLogFileModels    []SyncLogFileModel `json:"delCerSyncLogFileModels"`

	AddMftSyncLogFileModels    []SyncLogFileModel `json:"addMftSyncLogFileModels"`
	UpdateMftSyncLogFileModels []SyncLogFileModel `json:"updateMftSyncLogFileModels"`
	DelMftSyncLogFileModels    []SyncLogFileModel `json:"delMftSyncLogFileModels"`

	AddCrlSyncLogFileModels    []SyncLogFileModel `json:"addCrlSyncLogFileModels"`
	UpdateCrlSyncLogFileModels []SyncLogFileModel `json:"updateCrlSyncLogFileModels"`
	DelCrlSyncLogFileModels    []SyncLogFileModel `json:"delCrlSyncLogFileModels"`

	AddRoaSyncLogFileModels    []SyncLogFileModel `json:"addRoaSyncLogFileModels"`
	UpdateRoaSyncLogFileModels []SyncLogFileModel `json:"updateRoaSyncLogFileModels"`
	DelRoaSyncLogFileModels    []SyncLogFileModel `json:"delRoaSyncLogFileModels"`
}

func NewSyncLogFileModels(syncLogId uint64, dbSyncLogFileModels []SyncLogFileModel) *SyncLogFileModels {
	syncLogFileModels := &SyncLogFileModels{}
	syncLogFileModels.SyncLogId = syncLogId

	syncLogFileModels.AddCerSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.UpdateCerSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.DelCerSyncLogFileModels = make([]SyncLogFileModel, 0)

	syncLogFileModels.AddMftSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.UpdateMftSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.DelMftSyncLogFileModels = make([]SyncLogFileModel, 0)

	syncLogFileModels.AddCrlSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.UpdateCrlSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.DelCrlSyncLogFileModels = make([]SyncLogFileModel, 0)

	syncLogFileModels.AddRoaSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.UpdateRoaSyncLogFileModels = make([]SyncLogFileModel, 0)
	syncLogFileModels.DelRoaSyncLogFileModels = make([]SyncLogFileModel, 0)

	for i := range dbSyncLogFileModels {
		if dbSyncLogFileModels[i].FileType == "cer" {
			if dbSyncLogFileModels[i].SyncType == "add" {
				syncLogFileModels.AddCerSyncLogFileModels = append(syncLogFileModels.AddCerSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateCerSyncLogFileModels = append(syncLogFileModels.UpdateCerSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelCerSyncLogFileModels = append(syncLogFileModels.DelCerSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "mft" {
			if dbSyncLogFileModels[i].SyncType == "add" {
				syncLogFileModels.AddMftSyncLogFileModels = append(syncLogFileModels.AddMftSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateMftSyncLogFileModels = append(syncLogFileModels.UpdateMftSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelMftSyncLogFileModels = append(syncLogFileModels.DelMftSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "crl" {
			if dbSyncLogFileModels[i].SyncType == "add" {
				syncLogFileModels.AddCrlSyncLogFileModels = append(syncLogFileModels.AddCrlSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateCrlSyncLogFileModels = append(syncLogFileModels.UpdateCrlSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelCrlSyncLogFileModels = append(syncLogFileModels.DelCrlSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "roa" {
			if dbSyncLogFileModels[i].SyncType == "add" {
				syncLogFileModels.AddRoaSyncLogFileModels = append(syncLogFileModels.AddRoaSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateRoaSyncLogFileModels = append(syncLogFileModels.UpdateRoaSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelRoaSyncLogFileModels = append(syncLogFileModels.DelRoaSyncLogFileModels, dbSyncLogFileModels[i])
			}
		}
	}
	return syncLogFileModels
}
