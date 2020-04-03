package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// asn may be 0 ,or is nil. so use JsonInt
type SlurmAsnModel struct {
	Value    uint64
	IsNotNil bool
}

func (i *SlurmAsnModel) UnmarshalJSON(data []byte) error {
	// If this method was called, the value was set.
	i.IsNotNil = true

	// The key isn't set to null
	var temp uint64
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	i.Value = temp
	return nil
}
func (i *SlurmAsnModel) SqlNullInt() sql.NullInt64 {
	if !i.IsNotNil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: int64(i.Value),
		Valid: true,
	}
}

//////////////////////////////////////
//  SLURM
//////////////////
//lab_rpki_slurm
type SlurmModel struct {
	version uint64 `json:"slurmVersion" xorm:"version int"`
	//prefixFilter/bgpsecFilter/prefixAssertion/bgpsecAssertion',
	Style string        `json:"style" xorm:"style varchar(128)"`
	Asn   SlurmAsnModel `json:"asn" xorm:"asn int"`
	//198.51.100.0/24 or 2001:DB8::/32
	AddressPrefix string `json:"addressPrefix" xorm:"addressPrefix varchar(64)"`
	MaxLength     uint64 `json:"maxLength" xorm:"maxLength int"`
	//some base64 ski'
	Ski string `json:"ski" xorm:"ski varchar(256)"`
	//some base64 RouterPublicKey'
	RouterPublicKey string `json:"routerPublicKey" xorm:"routerPublicKey varchar(256)"`
	Comment         string `json:"comment" xorm:"comment varchar(256)"`
	//lab_rpki_slurm_file.id
	SlurmFileId uint64 `json:"slurmFileId" xorm:"slurmFileId  int"`
	//slurm_file.Priority '
	Priority uint64 `json:"priority" xorm:"priority  int"`
	//using/unused
	State string `json:"state" xorm:"state varchar(16)"`
}

//lab_rpki_slurm_file
type SlurmFileModel struct {
	FilePath   string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName   string    `json:"fileName" xorm:"fileName varchar(128)"`
	JsonAll    string    `json:"jsonAll" xorm:"jsonAll json"`
	UploadTime time.Time `json:"uploadTime" xorm:"uploadTime datetime"`
	//0-10, 0 is highest level, 10 is  lowest. default 5. the higher level user`s slurm will conver lower '
	Priority uint64 `json:"priority" xorm:"priority  int"`
}

// filter
// FormatPrefix, MaxPrefixLength and PrefixLength are not in json
// set asn==-1 means asn is empty
type PrefixFilters struct {
	Asn             SlurmAsnModel `json:"asn"`
	Prefix          string        `json:"prefix"`
	FormatPrefix    string        `json:"-"`
	PrefixLength    uint64        `json:"-"`
	MaxPrefixLength uint64        `json:"-"`
	Comment         string        `json:"comment"`
}

// set asn==-1 means asn is empty
type BgpsecFilters struct {
	Asn     SlurmAsnModel `json:"asn"`
	SKI     string        `json:"SKI"`
	Comment string        `json:"comment"`
}

type ValidationOutputFilters struct {
	PrefixFilters []PrefixFilters `json:"prefixFilters"`
	BgpsecFilters []BgpsecFilters `json:"bgpsecFilters"`
}

// assertion
// FormatPrefix, MaxPrefixLength and PrefixLength are not in json
// set asn==-1 means asn is empty
type PrefixAssertions struct {
	Asn             SlurmAsnModel `json:"asn"`
	Prefix          string        `json:"prefix"`
	MaxPrefixLength uint64        `json:"maxPrefixLength"`
	Comment         string        `json:"comment"`
}

// set asn==-1 means asn is empty
type BgpsecAssertions struct {
	Asn             SlurmAsnModel `json:"asn"`
	Comment         string        `json:"comment"`
	SKI             string        `json:"SKI"`
	RouterPublicKey string        `json:"RouterPublicKey"`
}

type LocallyAddedAssertions struct {
	PrefixAssertions []PrefixAssertions `json:"prefixAssertions"`
	BgpsecAssertions []BgpsecAssertions `json:"bgpsecAssertions"`
}

type Slurm struct {
	SlurmVersion            int                     `json:"slurmVersion"`
	ValidationOutputFilters ValidationOutputFilters `json:"validationOutputFilters"`
	LocallyAddedAssertions  LocallyAddedAssertions  `json:"locallyAddedAssertions"`
}
