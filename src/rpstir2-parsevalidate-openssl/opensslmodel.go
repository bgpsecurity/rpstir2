package openssl

//  1.2.840.113549.1.9.16.1.49
type AsProviderAttestation struct {
	AddressFamilyIdentifier Afi           `json:"addressFamilyIdentifier" asn1:"optional"`
	CustomerAsn             int           `json:"customerAsn"`
	ProviderAsns            []ProviderAsn `json:"ProviderAsns"`
}
type ProviderAsn struct {
	ProviderAsn             int `json:"providerAsn"`
	AddressFamilyIdentifier Afi `json:"addressFamilyIdentifier" asn1:"optional"`
}
type Afi []byte
