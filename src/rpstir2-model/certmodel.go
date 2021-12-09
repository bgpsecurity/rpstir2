package model

import (
	"crypto/x509/pkix"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// CER
type CerModel struct {
	Sn        string    `json:"sn" xorm:"sn varchar(128)"`
	NotBefore time.Time `json:"notBefore" xorm:"notBefore datetime"`
	NotAfter  time.Time `json:"notAfter" xorm:"notAfter datetime"`
	Subject   string    `json:"subject" xorm:"subject varchar(512)"`
	Issuer    string    `json:"issuer" xorm:"issuer  varchar(512)"`
	Ski       string    `json:"ski" xorm:"ski varchar(128)"`
	Aki       string    `json:"aki" xorm:"aki varchar(128)"`
	FilePath  string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName  string    `json:"fileName" xorm:"fileName varchar(128)"`
	FileHash  string    `json:"fileHash" xorm:"fileHash varchar(512)"`

	Version               int                   `json:"version"`
	BasicConstraintsModel BasicConstraintsModel `json:"basicConstraintsModel"`
	IsCa                  bool                  `json:"isCa"`
	IsRoot                bool                  `json:"isRoot"` //should have 5 root cer files
	SubjectAll            string                `json:"subjectAll"`
	IssuerAll             string                `json:"issuerAll"`
	KeyUsageModel         KeyUsageModel         `json:"keyUsageModel"`
	ExtKeyUsages          []int                 `json:"extKeyUsages"`
	//SHA256WithRSAEncryption
	SignatureInnerAlgorithm Sha256RsaModel `json:"signatureInnerAlgorithm"`
	//SHA256WithRSAEncryption
	SignatureOuterAlgorithm Sha256RsaModel `json:"signatureOuterAlgorithm"`
	//SHA256WithRSAEncryption
	PublicKeyAlgorithm RsaModel `json:"publicKeyAlgorithm"`

	CrldpModel        CrldpModel        `json:"crldpModel"`
	CerIpAddressModel CerIpAddressModel `json:"cerIpAddressModel"`
	AsnModel          AsnModel          `json:"asnModel"`
	AiaModel          AiaModel          `json:"aiaModel"`
	SiaModel          SiaModel          `json:"siaModel"`
	CertPolicyModel   CertPolicyModel   `json:"certPolicyModel"`
	ExtensionModels   []ExtensionModel  `json:"extensionModels"`
}

// IPAddress
type CerIpAddressModel struct {
	CerIpAddresses []CerIpAddress `json:"cerIpAddresses"`
	Critical       bool           `json:"critical"`
}
type CerIpAddress struct {
	AddressFamily uint64 `json:"addressFamily"`
	//address prefix: 147.28.83.0/24 '
	AddressPrefix string `json:"addressPrefix" xorm:"addressPrefix varchar(512)"`
	//min address:  99.96.0.0
	Min string `json:"min" xorm:"min varchar(512)"`
	//max address:   99.105.127.255
	Max string `json:"max" xorm:"max varchar(512)"`
	//min address range from addressPrefix or min/max, in hex:  63.60.00.00'
	RangeStart string `json:"rangeStart" xorm:"rangeStart varchar(512)"`
	//max address range from addressPrefix or min/max, in hex:  63.69.7f.ff'
	RangeEnd string `json:"rangeEnd" xorm:"rangeEnd varchar(512)"`
	//min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}
	AddressPrefixRange string `json:"addressPrefixRange" xorm:"addressPrefixRange json"`
}

// Asn
type AsnModel struct {
	Asns     []Asn `json:"asns"`
	Critical bool  `json:"critical"`
}

// asn, min, max default is -1
type Asn struct {
	Asn int64 `json:"asn" xorm:"asn bigint"`
	Min int64 `json:"min" xorm:"min bigint"`
	Max int64 `json:"max" xorm:"max bigint"`
}

func NewAsn() Asn {
	asn := Asn{
		Asn: -1,
		Min: -1,
		Max: -1,
	}
	return asn
}

// AIA
type AiaModel struct {
	CaIssuers string `json:"caIssuers" xorm:"caIssuers varchar(512)"`
	Critical  bool   `json:"critical"`
}

// SIA
type SiaModel struct {
	RpkiManifest string `json:"rpkiManifest" xorm:"rpkiManifest varchar(512)"`
	RpkiNotify   string `json:"rpkiNotify" xorm:"rpkiNotify varchar(512)"`
	CaRepository string `json:"caRepository" xorm:"caRepository varchar(512)"`
	SignedObject string `json:"signedObject" xorm:"signedObject varchar(512)"`
	Critical     bool   `json:"critical"`
}

// Crldp
type CrldpModel struct {
	Crldps   []string `json:"crldps" xorm:"crldps varchar(512)"`
	Critical bool     `json:"critical"`
}

// sha256WithRSAEncryption
// rsaEncryption
type RsaModel struct {
	Name string `json:"name"`
	// "85:89:43:5d:71:af:...."
	Modulus  string `json:"modulus"`
	Exponent uint64 `json:"exponent"`
}
type Sha256RsaModel struct {
	Name string `json:"name"`
	// may empty
	// "85:89:43:5d:71:af:...."
	Sha256 string `json:"sha256"`
}

/*
x509
const (
	KeyUsageDigitalSignature KeyUsage = 1 << iota
	KeyUsageContentCommitment
	KeyUsageKeyEncipherment
	KeyUsageDataEncipherment
	KeyUsageKeyAgreement
	KeyUsageCertSign
	KeyUsageCRLSign
	KeyUsageEncipherOnly
	KeyUsageDecipherOnly
)
*/
type KeyUsageModel struct {
	KeyUsage      int    `json:"keyUsage"`
	Critical      bool   `json:"critical"`
	KeyUsageValue string `json:"keyUsageValue"`
}

// certPolicy
type CertPolicyModel struct {
	Critical bool   `json:"critical"`
	Cps      string `json:"cps"`
}

// basic constraints
type BasicConstraintsModel struct {
	BasicConstraintsValid bool `json:"basicConstraintsValid"`
	Critical              bool `json:"critical"`
}

/*
   static char const *const allowed_extensions[] = {
        id_basicConstraints,
        id_subjectKeyIdentifier,
        id_authKeyId,
        id_keyUsage,
        id_extKeyUsage,         // allowed in future BGPSEC EE certs
        id_cRLDistributionPoints,
        id_pkix_authorityInfoAccess,
        id_pe_subjectInfoAccess,
        id_certificatePolicies,
        id_pe_ipAddrBlock,
        id_pe_autonomousSysNum,
        NULL
    };
*/
var CerExtensionOids = map[string]string{
	"2.5.29.14":          "Ski",              //subjectKeyIdentifier
	"2.5.29.35":          "Aki",              //authorityKeyIdentifier
	"2.5.29.15":          "KeyUsage",         //keyUsage
	"2.5.29.19":          "basicConstraints", //basicConstraints
	"2.5.29.31":          "Crldp",            //CRL Distribution Points
	"1.3.6.1.5.5.7.1.1":  "Aia",              // Authority Information Access
	"1.3.6.1.5.5.7.1.11": "Sia",              //Subject Information Access
	"2.5.29.32":          "CertPolicy",       //Certificate Policies
	"1.3.6.1.5.5.7.1.7":  "CerIpAddress",     //sbgp-ipAddrBlock
	"1.3.6.1.5.5.7.1.8":  "Asn",              //  sbgp-autonomousSysNum
}

// extensionModel
type ExtensionModel struct {
	Oid      string `json:"oid"`
	Critical bool   `json:"critical"`
	Name     string `json:"name"`
}

// sia.rpkiNotify/caRepository,
// subjectPublicKeyInfo
type ParseCerSimple struct {
	RpkiNotify           string `json:"rpkiNotify"`
	CaRepository         string `json:"caRepository"`
	SubjectPublicKeyInfo []byte `json:"subjectPublicKeyInfo"`
}

type SyncStyle struct {
	// sync/rrdp/rsync
	SyncStyle string `json:"syncStyle"`
}

///////////////////////////////////////////////////////////////////////////////
// CRL
type CrlModel struct {
	ThisUpdate time.Time `json:"thisUpdate" xorm:"thisUpdate datetime"`
	NextUpdate time.Time `json:"nextUpdate" xorm:"nextUpdate  datetime"`
	HasExpired string    `json:"hasExpired" xorm:"hasExpired  varchar(8)"`
	Aki        string    `json:"aki" xorm:"aki varchar(128)"`
	CrlNumber  uint64    `json:"crlNumber" xorm:"crlNumber int unsigned"`
	FilePath   string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName   string    `json:"fileName" xorm:"fileName varchar(128)"`
	FileHash   string    `json:"fileHash" xorm:"fileHash varchar(512)"`

	Version   int    `json:"version"`
	IssuerAll string `json:"issuerAll"`

	//AlgorithmIdentifier sha256withRSAEncryption
	/*Certificate  ::=  SEQUENCE  {
	     signatureAlgorithm   AlgorithmIdentifier,
	  TBSCertificate  ::=  SEQUENCE  {
	     signature            AlgorithmIdentifier,
	*/
	CertAlgorithm string `json:"certAlgorithm"`
	TbsAlgorithm  string `json:"tbsAlgorithm"`

	RevokedCertModels []RevokedCertModel `json:"revokedCertModels"`
}

type RevokedCertModel struct {
	Sn             string           `json:"sn" xorm:"sn varchar(512)"`
	RevocationTime time.Time        `json:"revocationTime" xorm:"revocationTime dateimte"`
	Extensions     []pkix.Extension `json:"extensions"`
}

///////////////////////////////////////////////////////////////////////////////
// MFT
// in some mft, manifestNumber is too too too too large, ft, so have to use string
type MftModel struct {

	// must be 0, or no in file
	//The version number of this version of the manifest specification MUST be 0.
	Version int `json:"version"`
	// have too big number, so using string
	MftNumber  string    `json:"mftNumber" xorm:"mftNumber varchar(1024)"`
	ThisUpdate time.Time `json:"thisUpdate" xorm:"thisUpdate datetime"`
	NextUpdate time.Time `json:"nextUpdate" xorm:"nextUpdate  datetime"`
	Ski        string    `json:"ski" xorm:"ski varchar(128)"`
	Aki        string    `json:"aki" xorm:"aki varchar(128)"`
	FilePath   string    `json:"filePath" xorm:"filePath varchar(512)"`
	FileName   string    `json:"fileName" xorm:"fileName varchar(128)"`
	FileHash   string    `json:"fileHash" xorm:"fileHash varchar(512)"`

	//OID: 1.2.840.113549.1.9.16.1.26
	EContentType string `json:"eContentType"`

	FileHashAlg    string          `json:"fileHashAlg"`
	FileHashModels []FileHashModel `json:"fileHashModels"`
	SiaModel       SiaModel        `json:"siaModel"`
	AiaModel       AiaModel        `json:"aiaModel"`

	EeCertModel     EeCertModel     `json:"eeCertModel"`
	SignerInfoModel SignerInfoModel `json:"signerInfoModel"`
}

type FileHashModel struct {
	File string `json:"file" xorm:"file varchar(1024)"`
	Hash string `json:"hash" xorm:"file varchar(1024)"`
}

////////////////////////////////////////
// Roa
type RoaModel struct {
	// must be 0, but always is not in file actually
	//The version number of this version of the roa specification MUST be 0.
	Version int `json:"version"`

	Asn      int64  `json:"asn" xorm:"asn bigint"`
	Ski      string `json:"ski" xorm:"ski varchar(128)"`
	Aki      string `json:"aki" xorm:"aki varchar(128)"`
	FilePath string `json:"filePath" xorm:"filePath varchar(512)"`
	FileName string `json:"fileName" xorm:"fileName varchar(128)"`
	FileHash string `json:"fileHash" xorm:"fileHash varchar(512)"`

	//OID: 1.2.240.113549.1.9.16.1.24
	EContentType string `json:"eContentType"`

	RoaIpAddressModels []RoaIpAddressModel `json:"roaIpAddressModels"`
	SiaModel           SiaModel            `json:"siaModel"`
	AiaModel           AiaModel            `json:"aiaModel"`

	EeCertModel     EeCertModel     `json:"eeCertModel"`
	SignerInfoModel SignerInfoModel `json:"signerInfoModel"`
}

type RoaIpAddressModel struct {
	AddressFamily uint64 `json:"addressFamily" xorm:"addressFamily int unsigned"`
	AddressPrefix string `json:"addressPrefix" xorm:"addressPrefix varchar(512)"`
	MaxLength     uint64 `json:"maxLength" xorm:"maxLength int unsigned"`
	//min address range from addressPrefix or min/max, in hex:  63.60.00.00'
	RangeStart string `json:"rangeStart" xorm:"rangeStart varchar(512)"`
	//max address range from addressPrefix or min/max, in hex:  63.69.7f.ff'
	RangeEnd string `json:"rangeEnd" xorm:"rangeEnd varchar(512)"`
	//min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}
	AddressPrefixRange string `json:"addressPrefixRange" xorm:"addressPrefixRange json"`
}

func (my *RoaModel) Compare(other *RoaModel) bool {
	if my.Asn != other.Asn {
		return false
	}
	if len(my.RoaIpAddressModels) != len(other.RoaIpAddressModels) {
		return false
	}
	for i := range my.RoaIpAddressModels {
		found := false
		for j := range other.RoaIpAddressModels {
			if my.RoaIpAddressModels[i].AddressFamily == other.RoaIpAddressModels[j].AddressFamily &&
				my.RoaIpAddressModels[i].AddressPrefix == other.RoaIpAddressModels[j].AddressPrefix &&
				my.RoaIpAddressModels[i].MaxLength == other.RoaIpAddressModels[j].MaxLength {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

///////////////////////////////////////////////
// EE
// EE in CerModel, MftModel, RoaModel, to get X509 Info and aia/sia/aki/ski
// https://datatracker.ietf.org/doc/rfc6488/?include_text=1
type EeCertModel struct {
	// must be 3
	Version int `json:"version"`
	// SHA256-RSA: x509.SignatureAlgorithm
	DigestAlgorithm string        `json:"digestAlgorithm"`
	Sn              string        `json:"sn"`
	NotBefore       time.Time     `json:"notBefore"`
	NotAfter        time.Time     `json:"notAfter"`
	KeyUsageModel   KeyUsageModel `json:"keyUsageModel"`
	ExtKeyUsages    []int         `json:"extKeyUsages"`

	BasicConstraintsValid bool `json:"basicConstraintsValid"`
	IsCa                  bool `json:"isCa"`

	SubjectAll string `json:"subjectAll"`
	IssuerAll  string `json:"issuerAll"`

	SiaModel SiaModel `json:"siaModel"`
	// in roa, ee cert also has ip address
	CerIpAddressModel CerIpAddressModel `json:"cerIpAddressModel"`

	CrldpModel CrldpModel `json:"crldpModel"`

	// in mft/roa , eecert start/end byte:
	// eeCertByte:=OraByte[EeCertStart:EeCertEnd] eeCertByte:=MftByte[EeCertStart:EeCertEnd]
	EeCertStart uint64 `json:"eeCertStart"`
	EeCertEnd   uint64 `json:"eeCertEnd"`
}

/* rfc5280
KeyUsage ::= BIT STRING {
   digitalSignature        (0),
   nonRepudiation          (1),  -- recent editions of X.509 have
                              -- renamed this bit to contentCommitment
   keyEncipherment         (2),
   dataEncipherment        (3),
   keyAgreement            (4),
   keyCertSign             (5),
   cRLSign                 (6),
   encipherOnly            (7),
   decipherOnly            (8) }
*/

// https://datatracker.ietf.org/doc/rfc6488/?include_text=1
type SignerInfoModel struct {
	// must be 3
	Version int `json:"version"`
	// 2.16.840.1.101.3.4.2.1 sha-256, must be sha256
	DigestAlgorithm string `json:"digestAlgorithm"`

	// 1.2.840.113549.1.9.3 --> roa:1.2.840.113549.1.9.16.1.24  mft:1.2.840.113549.1.9.16.1.26
	ContentType string `json:"contentType"`
	// 1.2.840.113549.1.9.5
	SigningTime time.Time `json:"signingTime"`
	// 1.2.840.113549.1.9.4
	MessageDigest string `json:"messageDigest"`
}
