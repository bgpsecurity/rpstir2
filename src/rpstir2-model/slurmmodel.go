package model

import (
	"github.com/guregu/null"
)

//////////////////////////////////////
//  SLURM
//////////////////

// filter
// FormatPrefix, MaxPrefixLength and PrefixLength are not in json
// TreatLevel: critical/major/normal
type PrefixFilters struct {
	Asn             null.Int `json:"asn"`
	Prefix          string   `json:"prefix"`
	FormatPrefix    string   `json:"-"`
	PrefixLength    uint64   `json:"-"`
	MaxPrefixLength uint64   `json:"maxPrefixLength"`
	Comment         string   `json:"comment"`
	TreatLevel      string   `json:"treatLevel,omitempty"`
}

// set asn==-1 means asn is empty
type BgpsecFilters struct {
	Asn     null.Int `json:"asn"`
	SKI     string   `json:"SKI"`
	Comment string   `json:"comment"`
}

type AspaFilters struct {
	CustomerAsn  null.Int       `json:"customerAsid"`
	ProviderAsns []ProviderAsns `json:"providers"`
	Comment      string         `json:"comment"`
}

type ProviderAsns struct {
	ProviderAsn   null.Int `json:"providerAsid"`
	AddressFamily string   `json:"afiLimit"` //IPv4 IPV6
}

const (
	SLURM_PROVIDER_ASNS_ADDRESS_FAMILY_IPV4 = "IPv4"
	SLURM_PROVIDER_ASNS_ADDRESS_FAMILY_IPV6 = "IPv6"
)

type ValidationOutputFilters struct {
	PrefixFilters []PrefixFilters `json:"prefixFilters"`
	BgpsecFilters []BgpsecFilters `json:"bgpsecFilters"`
	AspaFilters   []AspaFilters   `json:"aspaFilters"`
}

// assertion
// set !asn.Valid means asn is empty
// TreatLevel: critical/major/normal
type PrefixAssertions struct {
	Asn             null.Int `json:"asn"`
	Prefix          string   `json:"prefix"`
	MaxPrefixLength uint64   `json:"maxPrefixLength"`
	Comment         string   `json:"comment"`
	TreatLevel      string   `json:"treatLevel"`
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
	AspaAssertions   []AspaAssertions   `json:"aspaAssertions"`
}

type AspaAssertions struct {
	CustomerAsn  null.Int       `json:"customerAsid"`
	ProviderAsns []ProviderAsns `json:"providers"`
	Comment      string         `json:"comment"`
}

type Slurm struct {
	SlurmVersion            int                     `json:"slurmVersion"`
	ValidationOutputFilters ValidationOutputFilters `json:"validationOutputFilters"`
	LocallyAddedAssertions  LocallyAddedAssertions  `json:"locallyAddedAssertions"`
}
