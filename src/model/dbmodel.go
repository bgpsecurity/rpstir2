package model

import (
	"time"

	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null/v6"
)

// ////////////////
// CER
// ////////////////
// lab_rpki_cer
type LabRpkiCer struct {
	CerModel

	Id            uint64    `json:"id" xorm:"id int"`
	JsonAll       string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId     uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64    `json:"syncLogFileId" xorm:"syncLogFileId int"`
	UpdateTime    time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

// lab_rpki_cer_ipaddress
type LabRpkiCerIpaddress struct {
	CerIpAddressModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

// lab_rpki_cer_asn
type LabRpkiCerAsn struct {
	AsnModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

// lab_rpki_cer_sia
type LabRpkiCerSia struct {
	SiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

// lab_rpki_cer_aia
type LabRpkiCerAia struct {
	AiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

// lab_rpki_cer_crldp
type LabRpkiCerCrldp struct {
	CrldpModel

	Id    uint64 `json:"id" xorm:"id int"`
	CerId uint64 `json:"cerId" xorm:"cerId int"`
}

// ////////////////
// CRL
// ////////////////
// lab_rpki_crl
type LabRpkiCrl struct {
	CrlModel

	Id            uint64    `json:"id" xorm:"id int"`
	JsonAll       string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId     uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64    `json:"syncLogFileId" xorm:"syncLogFileId int"`
	UpdateTime    time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

// lab_rpki_crl_revoked_cert
type LabRpkiCrlRevokedCert struct {
	RevokedCertModel

	Id    uint64 `json:"id" xorm:"id int"`
	CrlId uint64 `json:"crlId" xorm:"crlId int"`
}

// ////////////////
// MFT
// ////////////////
// lab_rpki_Mft
type LabRpkiMft struct {
	MftModel

	Id            uint64    `json:"id" xorm:"id int"`
	JsonAll       string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId     uint64    `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64    `json:"syncLogFileId" xorm:"syncLogFileId int"`
	UpdateTime    time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

// lab_rpki_mft_sia
type LabRpkiMftSia struct {
	SiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	MftId uint64 `json:"mftId" xorm:"mftId  int"`
}

// lab_rpki_mft_aia
type LabRpkiMftAia struct {
	AiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	MftId uint64 `json:"mftId" xorm:"mftId  int"`
}

// lab_rpki_mft_file_hash struct
type LabRpkiMftFileHash struct {
	FileHashModel

	Id    uint64 `json:"id" xorm:"id int"`
	MftId uint64 `json:"mftId" xorm:"mftId  int"`
}

// ////////////////
// ROA
// ////////////////
// lab_rpki_roa
type LabRpkiRoa struct {
	RoaModel

	Id         uint64    `json:"id" xorm:"id int"`
	JsonAll    string    `json:"jsonAll" xorm:"jsonAll json"`
	SyncLogId  uint64    `json:"syncLogId" xorm:"syncLogId int"`
	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`
}

// lab_rpki_roa_sia
type LabRpkiRoaSia struct {
	SiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	RoaId uint64 `json:"roaId" xorm:"roaId int"`
}

// lab_rpki_roa_aiastruct
type LabRpkiRoaAia struct {
	AiaModel

	Id    uint64 `json:"id" xorm:"id int"`
	RoaId uint64 `json:"roaId" xorm:"roaId int"`
}

// lab_rpki_roa_ipaddress
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

// lab_rpki_sync_log_file
type LabRpkiSyncLogFile struct {
	Id        uint64    `json:"id" xorm:"pk autoincr"`
	SyncLogId uint64    `json:"syncLogId" xorm:"syncLogId int"`
	FilePath  string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName  string    `json:"fileName" xorm:"fileName varchar(128)"`
	FileType  string    `json:"fileType" xorm:"fileType varchar(16)"` //cer/roa/mft/crl/asa, not dot
	SyncType  string    `json:"syncType" xorm:"syncType varchar(16)"` //add/update/del
	SyncTime  time.Time `json:"syncTime" xorm:"syncTime datetime"`    //sync time for every file
	SourceUrl string    `json:"sourceUrl" xorm:"sourceUrl varchar(512)"`
	JsonAll   string    `json:"jsonAll" xorm:"jsonAll json"`
	FileHash  string    `json:"fileHash" xorm:"fileHash varchar(512)"`
	SyncStyle string    `json:"syncStyle" xorm:"syncStyle varchar(16)"` //rrdp/rsync
	State     string    `json:"state" xorm:"state json"`                //LabRpkiSyncLogFileState:
}

func (c LabRpkiSyncLogFile) SimpleString() string {
	m := make(map[string]interface{})
	m["id"] = c.Id
	m["syncLogId"] = c.SyncLogId
	m["filePath"] = c.FilePath
	m["fileName"] = c.FileName
	m["fileType"] = c.FileType
	m["syncType"] = c.SyncType
	m["sourceUrl"] = c.SourceUrl
	m["syncTime"] = c.SyncTime.Format("2006-01-02 15:04:05")
	return jsonutil.MarshalJson(m)
}

type LabRpkiSyncLogFileState struct {
	//finished
	Sync string `json:"sync"`
	//notYet/finished
	UpdateCertTable string `json:"updateCertTable"`
	//notYet/finished
	Rtr string `json:"rtr"`
}

// include CertModel/StateModel/OriginModel
type SyncLogFileModel struct {
	Id        uint64 `json:"id" xorm:"pk autoincr"`
	SyncLogId uint64 `json:"syncLogId" xorm:"syncLogId int"`
	FilePath  string `json:"filePath" xorm:"filePath varchar(512)"`
	FileName  string `json:"fileName" xorm:"fileName varchar(128)"`
	FileType  string `json:"fileType" xorm:"fileType varchar(16)"`
	SyncType  string `json:"syncType" xorm:"syncType varchar(16)"` //add/update/del

	CertModel   interface{} `json:"certModel"`
	StateModel  StateModel  `json:"stateModel"`
	OriginModel OriginModel `json:"originModel"`
	CertId      uint64      `json:"certId" xorm:"certId int"` //cerId / mftId / roaId / crlId / asaId
	Index       uint64      `json:"index"`                    // db.rows index
}

func (c SyncLogFileModel) String() string {
	m := make(map[string]interface{})
	m["id"] = c.Id
	m["syncLogId"] = c.SyncLogId
	m["filePath"] = c.FilePath
	m["fileName"] = c.FileName
	m["fileType"] = c.FileType
	m["syncType"] = c.SyncType
	m["certId"] = c.CertId
	m["index"] = c.Index
	return jsonutil.MarshalJson(m)
}

// ////////////////
// RTR
// ////////////////
// lab_rpki_rtr_session
type LabRpkiRtrSession struct {
	//sessionId, after init will not change'
	SessionId  uint64    `json:"sessionId" xorm:"sessionId  int"`
	CreateTime time.Time `json:"createTime" xorm:"createTime datetime"`
}

// lab_rpki_rtr_full
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

// lab_rpki_rtr_full_log
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

// lab_rpki_rtr_incremental
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

// /////////////////////////////////////
// ASA
// /////////////////////////////////////
// lab_rpki_rtr_asa_full
type LabRpkiRtrAsaFull struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber int"`
	CustomerAsn  uint64 `json:"customerAsn" xorm:"customerAsn int"`
	ProviderAsns string `json:"providerAsns" xorm:"providerAsns json"`
	SourceFrom   string `json:"sourceFrom" xorm:"sourceFrom json"`
}

type LabRpkiRtrAsaFullLog struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber int"`
	CustomerAsn  uint64 `json:"customerAsn" xorm:"customerAsn int"`
	ProviderAsns string `json:"providerAsns" xorm:"providerAsns json"`
	SourceFrom   string `json:"sourceFrom" xorm:"sourceFrom json"`
}

type AsaToRtrFullLog struct {
	AsaId         uint64 `json:"asaId" xorm:"asaId int"`
	CustomerAsn   uint64 `json:"customerAsn" xorm:"customerAsn int"`
	ProviderAsns  string `json:"providerAsns" xorm:"providerAsn json"`
	SyncLogId     uint64 `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId uint64 `json:"syncLogFileId" xorm:"syncLogFileId int"`
}

// lab_rpki_rtr_asa_incremental
type LabRpkiRtrAsaIncremental struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	//announce/withdraw, is 1/0 in protocol
	Style        string `json:"style" xorm:"style varchar(16)"`
	CustomerAsn  uint64 `json:"customerAsn" xorm:"customerAsn int"`
	ProviderAsns string `json:"providerAsns" xorm:"providerAsns json"`
	//'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}',
	SourceFrom string `json:"sourceFrom" xorm:"sourceFrom json"`
}

// /////////////////////////////////////
// MOA
// /////////////////////////////////////
// lab_rpki_rtr_moa_full
type LabRpkiRtrMoaFull struct {
	Id                uint64 `json:"id" xorm:"id int"`
	SerialNumber      uint64 `json:"serialNumber" xorm:"serialNumber int"`
	Ipv6MappingPrefix string `json:"ipv6MappingPrefix" xorm:"ipv6MappingPrefix varchar(128)"`
	Ipv4Prefixes      string `json:"ipv4Prefixes" xorm:"ipv4Prefixes json"`
	SourceFrom        string `json:"sourceFrom" xorm:"sourceFrom json"`
}

type LabRpkiRtrMoaFullLog struct {
	Id                uint64 `json:"id" xorm:"id int"`
	SerialNumber      uint64 `json:"serialNumber" xorm:"serialNumber int"`
	Ipv6MappingPrefix string `json:"ipv6MappingPrefix" xorm:"ipv6MappingPrefix varchar(128)"`
	Ipv4Prefixes      string `json:"ipv4Prefixes" xorm:"ipv4Prefixes json"`
	SourceFrom        string `json:"sourceFrom" xorm:"sourceFrom json"`
}

type MoaToRtrFullLog struct {
	MoaId             uint64 `json:"moaId" xorm:"moaId int"`
	Ipv6MappingPrefix string `json:"ipv6MappingPrefix" xorm:"ipv6MappingPrefix varchar(128)"`
	Ipv4Prefixes      string `json:"ipv4Prefixes" xorm:"ipv4Prefixes json"`
	SyncLogId         uint64 `json:"syncLogId" xorm:"syncLogId int"`
	SyncLogFileId     uint64 `json:"syncLogFileId" xorm:"syncLogFileId int"`
}

// lab_rpki_rtr_moa_incremental
type LabRpkiRtrMoaIncremental struct {
	Id           uint64 `json:"id" xorm:"id int"`
	SerialNumber uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	//announce/withdraw, is 1/0 in protocol
	Style             string `json:"style" xorm:"style varchar(16)"`
	Ipv6MappingPrefix string `json:"ipv6MappingPrefix" xorm:"ipv6MappingPrefix varchar(128)"`
	Ipv4Prefixes      string `json:"ipv4Prefixes" xorm:"ipv4Prefixes json"`
	SourceFrom        string `json:"sourceFrom" xorm:"sourceFrom json"`
}

// //////////////////////////
// source
// /////////////////////////
type LabRpkiRtrSourceFrom struct {
	// sync/slurm/rushtransfer
	Source           string `json:"source"`
	SyncLogId        uint64 `json:"syncLogId"`
	SyncLogFileId    uint64 `json:"syncLogFileId"`
	SlurmId          uint64 `json:"slurmId"`
	SlurmLogId       uint64 `json:"slurmLogId"`
	SlurmLogFileId   uint64 `json:"slurmLogFileId"`
	RushTransferUuid string `json:"transferUuid"`
}

//////////////////
//  SLURM
//////////////////

type SlurmToRtrFullLog struct {
	Id    uint64 `json:"id" xorm:"id int"`
	Style string `json:"style" xorm:"style varchar(128)"`

	Asn          null.Int `json:"asn" xorm:"asn int"`
	Address      string   `json:"address" xorm:"address varchar(256)"`
	PrefixLength uint64   `json:"prefixLength" xorm:"prefixLength int"`
	MaxLength    uint64   `json:"maxLength" xorm:"maxLength int"`

	CustomerAsn   null.Int `json:"customerAsn" xorm:"customerAsn int"`
	ProviderAsn   null.Int `json:"providerAsn" xorm:"providerAsn int"`
	AddressFamily string   `json:"addressFamily" xorm:"addressFamily varchar(16)"`

	SlurmId        uint64 `json:"slurmId" xorm:"slurmId int"`
	SlurmLogId     uint64 `json:"slurmLogId" xorm:"slurmLogId int"`
	SlurmLogFileId uint64 `json:"slurmLogFileId" xorm:"slurmLogFileId int"`
	SourceFromJson string `json:"sourceFromJson" xorm:"sourceFromJson json"`
}

// //////////////////////////////////
// rrdp
// /////////////////////////////////
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
