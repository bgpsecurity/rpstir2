package openssl

// 1.2.840.113549.1.9.16.1.49
type AsProviderAttestation struct {
	Version      Version `json:"version" asn1:"class:2,tag:0"` //default 0
	CustomerAsn  int     `json:"customerAsn"`
	ProviderAsns []int   `json:"providerAsns"`
}
type Version struct {
	Version int `json:"version"`
}

type AsProviderAttestationOld struct {
	//Version                 int           `json:"version" asn1:"optional"` //default 0
	AddressFamilyIdentifier Afi              `json:"addressFamilyIdentifier" asn1:"optional"`
	CustomerAsn             int              `json:"customerAsn"`
	ProviderAsns            []ProviderAsnOld `json:"providerAsns"`
}
type ProviderAsnOld struct {
	ProviderAsn             int `json:"providerAsn"`
	AddressFamilyIdentifier Afi `json:"addressFamilyIdentifier" asn1:"optional"`
}
type Afi []byte

func convertAsProviderAttestationOldToAsProviderAttestation(old AsProviderAttestationOld) AsProviderAttestation {
	as := AsProviderAttestation{}
	as.CustomerAsn = old.CustomerAsn
	as.ProviderAsns = make([]int, 0)
	for i := range old.ProviderAsns {
		providerAsn := old.ProviderAsns[i].ProviderAsn
		as.ProviderAsns = append(as.ProviderAsns, providerAsn)
	}
	return as
}
