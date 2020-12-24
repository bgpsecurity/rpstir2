package model

import (
	"database/sql"
	"time"
)

//////////////////
// CER
//////////////////
//lab_rpki_cer
type LabRpkiCer struct {
	CerModel

	Id            uint64    `json:"id" xorm:"id int"`
	JsonAll       string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId     uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64    `json:"syncLogFileId" xorm:"syncLogFileId int"`
	UpdateTime    time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

//lab_rpki_cer_ipaddress
type LabRpkiCerIpaddress struct {
	CerIpAddressModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

//lab_rpki_cer_asn
type LabRpkiCerAsn struct {
	AsnModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

//lab_rpki_cer_sia
type LabRpkiCerSia struct {
	SiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

//lab_rpki_cer_aia
type LabRpkiCerAia struct {
	AiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

//lab_rpki_cer_crldp
type LabRpkiCerCrldp struct {
	CrldpModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

//////////////////
// CRL
//////////////////
//lab_rpki_crl
type LabRpkiCrl struct {
	CrlModel

	Id            uint64    `json:"id" xorm:"id int"`
	JsonAll       string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId     uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64    `json:"syncLogFileId" xorm:"syncLogFileId int"`
	UpdateTime    time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

//lab_rpki_crl_revoked_cert
type LabRpkiCrlRevokedCert struct {
	RevokedCertModel

	Id    uint64 `json:"id" xorm:"id int"`
	CrlId uint64 `json:"crlId" xorm:"crlId int"`
}

//////////////////
// MFT
//////////////////
//lab_rpki_Mft
type LabRpkiMft struct {
	MftModel

	Id            uint64    `json:"id" xorm:"id int"`
	JsonAll       string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId     uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64    `json:"syncLogFileId" xorm:"syncLogFileId int"`
	UpdateTime    time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

//lab_rpki_mft_sia
type LabRpkiMftSia struct {
	SiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	MftId uint64 `json:"mftId" xorm:"mftId  int"`
}

//lab_rpki_mft_aia
type LabRpkiMftAia struct {
	AiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	MftId uint64 `json:"mftId" xorm:"mftId  int"`
}

//lab_rpki_mft_file_hash struct
type LabRpkiMftFileHash struct {
	FileHashModel

	Id    uint64 `json:"id" xorm:"id int"`
	MftId uint64 `json:"mftId" xorm:"mftId  int"`
}

//////////////////
// ROA
//////////////////
//lab_rpki_roa
type LabRpkiRoa struct {
	RoaModel

	Id         uint64    `json:"id" xorm:"id int"`
	JsonAll    string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId  uint64    `json:"syncLogId" xorm:"syncLogId int"`
	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

//lab_rpki_roa_sia
type LabRpkiRoaSia struct {
	SiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	RoaId uint64 `json:"roaId" xorm:"roaId int"`
}

//lab_rpki_roa_aiastruct
type LabRpkiRoaAia struct {
	AiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	RoaId uint64 `json:"roaId" xorm:"roaId int"`
}

//lab_rpki_roa_ipaddress
type LabRpkiRoaIpaddress struct {
	RoaIpAddressModel

	Id    uint64 `json:"id" xorm:"id int"`
	RoaId uint64 `json:"roaId" xorm:"roaId int"`
}

type LabRpkiRoaIpaddressView struct {
	Id            uint64 `json:"id" xorm:"id int"`
	Asn           int64  `json:"asn" xorm:"asn bigint"`
	AddressPrefix string `json:"addressPrefix" xorm:"addressPrefix varchar(512)"`
	MaxLength     uint64 `json:"maxLength" xorm:"maxLength int"`
	SyncLogId     uint64 `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64 `json:"syncLogFileId" xorm:"syncLogFileId int"`
}

//////////////////
// recored every sync log for cer/crl/roa/mft
//////////////////

// lab_rpki_sync_log
type LabRpkiSyncLog struct {
	Id uint64 `json:"id" xorm:"id"`

	//rsync/delta
	SyncStyle          string `json:"syncStyle" xorm:"syncStyle varchar(16)"`
	SyncState          string `json:"syncState" xorm:"syncState json"`
	ParseValidateState string `json:"parseValidateState" xorm:"parseValidateState json"`
	ChainValidateState string `json:"chainValidateState" xorm:"chainValidateState json"`
	RtrState           string `json:"rtrState" xorm:"rtrState json"`

	//rsyncing   diffing/diffed   parsevalidating/parsevalidated   rtring/rtred idle
	State string `json:"state" xorm:"state varchar(16)"`
}

//lab_rpki_sync_log_file
type LabRpkiSyncLogFile struct {
	Id        uint64 `json:"id" xorm:"pk autoincr"`
	SyncLogId uint64 `json:"syncLogId" xorm:"syncLogId int"`
	//cer/roa/mft/crl, not dot
	FileType string `json:"fileType" xorm:"fileType varchar(16)"`
	//sync time for every file
	SyncTime time.Time `json:"syncTime" xorm:"syncTime datetime"`
	FilePath string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName string    `json:"fileName" xorm:"fileName varchar(128)"`
	JsonAll  string    `json:"jsonAll" xorm:"jsonAll json"`
	FileHash string    `json:"fileHash" xorm:"fileHash varchar(512)"`
	//add/update/del
	SyncType string `json:"syncType" xorm:"syncType varchar(16)"`
	//rrdp/rsync
	SyncStyle string `json:"syncStyle" xorm:"syncStyle varchar(16)"`
	//LabRpkiSyncLogFileState:
	State string `json:"state" xorm:"state json"`
}

type LabRpkiSyncLogFileState struct {
	//finished
	Sync string `json:"sync"`
	//notYet/finished
	UpdateCertTable string `json:"updateCertTable"`
	//notYet/finished
	Rtr string `json:"rtr"`
}

//////////////////
// RTR
//////////////////
//lab_rpki_rtr_session
type LabRpkiRtrSession struct {
	//sessionId, after init will not change'
	SessionId  uint64    `json:"sessionId" xorm:"sessionId  int"`
	CreateTime time.Time `json:"createTime" xorm:"createTime datetime"`
}

//lab_rpki_rtr_serial_number
type LabRpkiRtrSerialNumber struct {
	Id           uint64    `json:"id" xorm:"id int"`
	SerialNumber uint64    `json:"serialNumber" xorm:"serialNumber bigint"`
	CreateTime   time.Time `json:"createTime" xorm:"createTime datetime"`
}

//lab_rpki_rtr_full
type LabRpkiRtrFull struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	Asn          int64  `json:"asn" xorm:"asn bigint"`
	//address: 63.60.00.00
	Address      string `json:"address" xorm:"address varchar(512)"`
	PrefixLength uint64 `json:"prefixLength" xorm:"prefixLength int"`
	MaxLength    uint64 `json:"maxLength" xorm:"maxLength int"`
	//'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}',
	SourceFrom string `json:"sourceFrom" xorm:"sourceFrom json"`
}

//lab_rpki_rtr_full_log
type LabRpkiRtrFullLog struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	Asn          int64  `json:"asn" xorm:"asn bigint"`
	//address: 63.60.00.00
	Address      string `json:"address" xorm:"address varchar(512)"`
	PrefixLength uint64 `json:"prefixLength" xorm:"prefixLength int"`
	MaxLength    uint64 `json:"maxLength" xorm:"maxLength int"`
	//'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}',
	SourceFrom string `json:"sourceFrom" xorm:"sourceFrom json"`
}

type RoaToRtrFullLog struct {
	RoaId         uint64 `json:"roaId" xorm:"roaId int"`
	Asn           int64  `json:"asn" xorm:"asn bigint"`
	Address       string `json:"address" xorm:"address  varchar(512)"`
	PrefixLength  uint64 `json:"prefixLength" xorm:"prefixLength int"`
	MaxLength     uint64 `json:"maxLength" xorm:"maxLength int"`
	SyncLogId     uint64 `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64 `json:"syncLogFileId" xorm:"syncLogFileId int"`
}

//lab_rpki_rtr_incremental
type LabRpkiRtrIncremental struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	//announce/withdraw, is 1/0 in protocol
	Style string `json:"style" xorm:"style varchar(16)"`
	Asn   int64  `json:"asn" xorm:"asn bigint"`
	//address: 63.60.00.00
	Address      string `json:"address" xorm:"address varchar(512)"`
	PrefixLength uint64 `json:"prefixLength" xorm:"prefixLength int"`
	MaxLength    uint64 `json:"maxLength" xorm:"maxLength int"`
	//'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}',
	SourceFrom string `json:"sourceFrom" xorm:"sourceFrom json"`
}

type LabRpkiRtrSourceFrom struct {
	Source        string `json:"source"`
	SyncLogId     uint64 `json:"syncLogId"`
	SyncLogFileId uint64 `json:"syncLogFileId"`
	SlurmId       uint64 `json:"slurmId"`
	SlurmFileId   uint64 `json:"slurmFileId"`
	TransferLogId uint64 `json:"transferLogId"`
}

//////////////////
//  SLURM
//////////////////

// because asn may be nil or be 0, so using  sql.NullInt64
type SlurmToRtrFullLog struct {
	Id           uint64        `json:"id" xorm:"id int"`
	Style        string        `json:"style" xorm:"style varchar(128)"`
	Asn          sql.NullInt64 `json:"asn" xorm:"asn int"`
	Address      string        `json:"address" xorm:"address varchar(256)"`
	PrefixLength uint64        `json:"prefixLength" xorm:"prefixLength int"`
	MaxLength    uint64        `json:"maxLength" xorm:"maxLength int"`
	SlurmFileId  uint64        `json:"slurmFileId" xorm:"slurmFileId int"`
	SlurmId      uint64        `json:"slurmId" xorm:"slurmId int"`
}

//lab_rpki_slurm_file
type LabRpkiSlurmFile struct {
	Id         uint64    `json:"id" xorm:"id int"`
	JsonAll    string    `json:"jsonAll" xorm:"jsonAll json"`
	UploadTime time.Time `json:"uploadTime" xorm:"uploadTime datetime"`
	FilePath   string    `json:"filePath" xorm:"filePath varchar(128)"`
	FileName   string    `json:"fileName" xorm:"fileName varchar(128)"`
}

//////////////////
//  rp transfer
//////////////////
//lab_rpki_transfer_target
type LabRpkiTransferTarget struct {
	Id uint64 `json:"id" xorm:"id int"`
	//http/https
	Protocol string `json:"protocol" xorm:"protocol varchar(64)"`
	//IP or domain
	Address string `json:"address" xorm:"address  varchar(64)"`
	Port    uint64 `json:"port" xorm:"port  int"`
	//vc/rp
	TargetType string `json:"targetType" xorm:"targetType varchar(64)"`
	//update time
	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`
	//valid/invalid
	State string `json:"state" xorm:"state varchar(64)"`
}

//lab_rpki_transfer_log
type LabRpkiTransferLog struct {
	Id uint64 `json:"id" xorm:"id int"`
	//lab_rpki_transfer_target.id
	TransferTargetId uint64 `json:"transferTargetId" xorm:"transferTargetId int"`
	//all/update',
	Operate string `json:"operate" xorm:"operate varchar(64)"`
	//transfer time
	TransferTime time.Time `json:"transferTime" xorm:"transferTime datetime"`
	Uuid         string    `json:"uuid" xorm:"uuid  varchar(64)"`
	Content      string    `json:"content" xorm:"content longtext"`
	//send/receive
	TransferType string `json:"transferType" xorm:"transferType varchar(64)"`
	//ok/fail
	Result string `json:"result" xorm:"result varchar(64)"`
	ErrMsg string `json:"errMsg" xorm:"errMsg varchar(256)"`
}
////////////////////////////////////
// rrdp
///////////////////////////////////
type LabRpkiSyncRrdpLog struct {
	Id         uint64    `json:"id" xorm:"id int"`
	SyncLogId  uint64    `json:"syncLogId" xorm:"syncLogId int"`
	NotifyUrl  string    `json:"notifyUrl" xorm:"notifyUrl varchar(512)"`
	SessionId  string    `json:"sessionId" xorm:"sessionId varchar(512)"`
	LastSerial uint64    `json:"lastSerial" xorm:"lastSerial int"`
	CurSerial  uint64    `json:"curSerial" xorm:"curSerial int"`
	RrdpTime   time.Time `json:"rrdpTime" xorm:"rrdpTime datetime"`
	//snapshot/delta
	RrdpType string `json:"rrdpType" xorm:"rrdpType varchar(16)"`
}
