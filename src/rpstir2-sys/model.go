package sys

type SysStyle struct {
	// "init" :  will create all table;
	// "fullsync": will remove current data to forece full sync data, and retain rtr/slurm/transfer data.
	// "resetall" will remove all data including rtr/slurm/transfer;
	SysStyle string `json:"sysStyle"`
	// direct/entire
	SyncPolicy string `json:"syncPolicy"`
}

type CertResults struct {
	CerResult CertResult `json:"cerResult"`
	CrlResult CertResult `json:"crlResult"`
	MftResult CertResult `json:"mftResult"`
	RoaResult CertResult `json:"roaResult"`
}

type CertResult struct {
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

//  {"prefix": "192.0.2.0/24", "asn": 0, "max_length": 24},
type RtrForManrs struct {
	Asn       int    `json:"asn" xorm:"asn int"`
	Prefix    string `json:"prefix"`
	MaxLength int    `json:"max_length"  xorm:"max_length int"`

	Address      string `json:"-" xorm:"address varchar(255)"`
	PrefixLength int    `json:"-"  xorm:"prefixLength int"`
}
