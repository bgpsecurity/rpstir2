package model

import (
	"github.com/guregu/null"
)

//////////////////////////////////////
//  SLURM
//////////////////

// filter
// FormatPrefix, MaxPrefixLength and PrefixLength are not in json
type PrefixFilters struct {
	Asn             null.Int `json:"asn"`
	Prefix          string   `json:"prefix"`
	FormatPrefix    string   `json:"-"`
	PrefixLength    uint64   `json:"-"`
	MaxPrefixLength uint64   `json:"maxPrefixLength"`
	Comment         string   `json:"comment"`
}

// set asn==-1 means asn is empty
type BgpsecFilters struct {
	Asn     null.Int `json:"asn"`
	SKI     string   `json:"SKI"`
	Comment string   `json:"comment"`
}

type ValidationOutputFilters struct {
	PrefixFilters []PrefixFilters `json:"prefixFilters"`
	BgpsecFilters []BgpsecFilters `json:"bgpsecFilters"`
}

// assertion
// FormatPrefix, MaxPrefixLength and PrefixLength are not in json
// set asn==-1 means asn is empty
type PrefixAssertions struct {
	Asn             null.Int `json:"asn"`
	Prefix          string   `json:"prefix"`
	MaxPrefixLength uint64   `json:"maxPrefixLength"`
	Comment         string   `json:"comment"`
}

// set asn==-1 means asn is empty
type BgpsecAssertions struct {
	Asn             null.Int `json:"asn"`
	Comment         string   `json:"comment"`
	SKI             string   `json:"SKI"`
	RouterPublicKey string   `json:"RouterPublicKey"`
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
