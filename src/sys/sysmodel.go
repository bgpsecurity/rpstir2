package sys

// SysStyle
type SysStyle struct {

	// SysStyle
	// "init" :  will create all table;
	// "fullsync": will remove current data to forece full sync data, and retain rtr/slurm/transfer data.
	SysStyle string `json:"sysStyle"`

	// SyncPolicy distributed/entire
	SyncPolicy string `json:"syncPolicy"`
}

// CertResults
type CertResults struct {
	CerResult CertResult `json:"cerResult"`
	CrlResult CertResult `json:"crlResult"`
	MftResult CertResult `json:"mftResult"`
	RoaResult CertResult `json:"roaResult"`
	AsaResult CertResult `json:"asaResult"`
	MoaResult CertResult `json:"moaResult"`
}

// CertResult
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

// {"prefix": "192.0.2.0/24", "asn": 0, "max_length": 24},
type RtrForManrs struct {
	Asn       int    `json:"asn" xorm:"asn int"`
	Prefix    string `json:"prefix"`
	MaxLength int    `json:"max_length"  xorm:"max_length int"`

	Address      string `json:"-" xorm:"address varchar(255)"`
	PrefixLength int    `json:"-"  xorm:"prefixLength int"`
}

type ChainRepos struct {
	Repos []string `json:"repos"`
}

type ChainCertId struct {
	Id         uint64 `json:"id" xorm:"id int"`
	ParentCers string `json:"parentCers" xorm:"parentCers  varchar(255)"`
	FileType   string `json:"fileType" xorm:"fileType  varchar(16)"`
}

type ParentCerId struct {
	Id uint64 `json:"id" xorm:"id int"`
}
type CertIdRepo struct {
	Id       uint64 `json:"id" xorm:"id int"`
	Repo     string `json:"repo" xorm:"repo  varchar(255)"`
	FileType string `json:"fileType" xorm:"fileType  varchar(16)"`
}
