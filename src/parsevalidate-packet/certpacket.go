package packet

import (
	"bytes"
)

var oid = map[string]string{
	"2.5.4.3":                    "CN",
	"2.5.4.4":                    "SN",
	"2.5.4.5":                    "serialNumber",
	"2.5.4.6":                    "C",
	"2.5.4.7":                    "L",
	"2.5.4.8":                    "ST",
	"2.5.4.9":                    "streetAddress",
	"2.5.4.10":                   "O",
	"2.5.4.11":                   "OU",
	"2.5.4.12":                   "title",
	"2.5.4.17":                   "postalCode",
	"2.5.4.42":                   "GN",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generationQualifier",
	"2.5.4.46":                   "dnQualifier",
	"2.5.4.65":                   "pseudonym",
	"0.9.2342.19200300.100.1.25": "DC",
	"1.2.840.113549.1.9.1":       "emailAddress",
	"0.9.2342.19200300.100.1.1":  "userid",
	"2.5.29.20":                  "CRL Number",
}

type OidPacket struct {
	Oid          string
	OidPacket    *Packet
	ParentPacket *Packet
}

type Packet struct {
	ClassType   uint8
	TagType     uint8
	Tag         uint8
	Value       interface{}
	Data        *bytes.Buffer
	Children    []*Packet
	Description string
	Parent      *Packet
}

const (
	TagEOC              = 0x00
	TagBoolean          = 0x01
	TagInteger          = 0x02
	TagBitString        = 0x03
	TagOctetString      = 0x04
	TagNULL             = 0x05
	TagObjectIdentifier = 0x06
	TagObjectDescriptor = 0x07
	TagExternal         = 0x08
	TagRealFloat        = 0x09
	TagEnumerated       = 0x0a
	TagEmbeddedPDV      = 0x0b
	TagUTF8String       = 0x0c
	TagRelativeOID      = 0x0d
	TagSequence         = 0x10
	TagSet              = 0x11
	TagNumericString    = 0x12
	TagPrintableString  = 0x13
	TagT61String        = 0x14
	TagVideotexString   = 0x15
	TagIA5String        = 0x16
	TagUTCTime          = 0x17
	TagGeneralizedTime  = 0x18
	TagGraphicString    = 0x19
	TagVisibleString    = 0x1a
	TagGeneralString    = 0x1b
	TagUniversalString  = 0x1c
	TagCharacterString  = 0x1d
	TagBMPString        = 0x1e
	TagBitmask          = 0x1f // xxx11111b

	//private
	TagAsNum = 0xa0
	TagRdi   = 0xa1
)

var TagMap = map[uint8]string{
	TagEOC:              "EOC (End-of-Content)",
	TagBoolean:          "Boolean",
	TagInteger:          "Integer",
	TagBitString:        "Bit String",
	TagOctetString:      "Octet String",
	TagNULL:             "NULL",
	TagObjectIdentifier: "Object Identifier",
	TagObjectDescriptor: "Object Descriptor",
	TagExternal:         "External",
	TagRealFloat:        "Real (float)",
	TagEnumerated:       "Enumerated",
	TagEmbeddedPDV:      "Embedded PDV",
	TagUTF8String:       "UTF8 String",
	TagRelativeOID:      "Relative-OID",
	TagSequence:         "Sequence and Sequence of",
	TagSet:              "Set and Set OF",
	TagNumericString:    "Numeric String",
	TagPrintableString:  "Printable String",
	TagT61String:        "T61 String",
	TagVideotexString:   "Videotex String",
	TagIA5String:        "IA5 String",
	TagUTCTime:          "UTC Time",
	TagGeneralizedTime:  "Generalized Time",
	TagGraphicString:    "Graphic String",
	TagVisibleString:    "Visible String",
	TagGeneralString:    "General String",
	TagUniversalString:  "Universal String",
	TagCharacterString:  "Character String",
	TagBMPString:        "BMP String",
	//private
	TagAsNum: "ASNum",
	TagRdi:   "Rdi",
}

//first of byte, the second bites： 00 is Universal, 01 is APPLICATION, 10 is context-specific, 11 is PRIVATE
//https://blog.csdn.net/sever2012/article/details/7698297
const (
	ClassUniversal   = 0   // 00xxxxxxb
	ClassApplication = 64  // 01xxxxxxb
	ClassContext     = 128 // 10xxxxxxb
	ClassPrivate     = 192 // 11xxxxxxb
	ClassBitmask     = 192 // 11xxxxxxb
)

var ClassMap = map[uint8]string{
	ClassUniversal:   "Universal",
	ClassApplication: "Application",
	ClassContext:     "Context",
	ClassPrivate:     "Private",
}

//the 5th bites is  primitive code or constructed code
//https://blog.csdn.net/sever2012/article/details/7698297
const (
	TypePrimative   = 0  // xx0xxxxxb
	TypeConstructed = 32 // xx1xxxxxb
	TypeBitmask     = 32 // xx1xxxxxb
)

var TypeMap = map[uint8]string{
	TypePrimative:   "Primative",
	TypeConstructed: "Constructed",
}

// when it is 0xa0, it will found 0x00 0x00 from last to head,ft
const (
	TypeLastIndex = 0xa0
)

const (
	ipv4    = 0x01
	ipv6    = 0x02
	ipv4len = 32
	ipv6len = 128
)

const (
	oidAuthorityInfoAccessKey = "1.3.6.1.5.5.7.1.1"  //AIA   authorityInfoAccess
	oidCaIssuersKey           = "1.3.6.1.5.5.7.48.2" // CAIssuers  url, in AIA

	oidSubjectInfoAccessKey = "1.3.6.1.5.5.7.1.11"  //SIA subjectInfoAccess
	oidRpkiManifestKey      = "1.3.6.1.5.5.7.48.10" // manifest url,  in SIA ,
	oidRpkiNotifyKey        = "1.3.6.1.5.5.7.48.13" // rpkiNotify url, in SIA
	oidCaRepositoryKey      = "1.3.6.1.5.5.7.48.5"  // caRepository url, in SIA
	oidSignedObjectKey      = "1.3.6.1.5.5.7.48.11" // signedObject rsync url, in SIA

	oidSubjectKeyIdentifierKey   = "2.5.29.14" //SKI ,  subjectKeyIdentifier
	oidAuthorityKeyIdentifierKey = "2.5.29.35" //AKI, authorityKeyIdentifier

	oidIpAddressKey = "1.3.6.1.5.5.7.1.7"
	oidASKey        = "1.3.6.1.5.5.7.1.8"
	oidManifestKey  = "1.2.840.113549.1.9.16.1.26"
	oidRoaKey       = "1.2.840.113549.1.9.16.1.24"

	newOidIpAddressKey = "1.3.6.1.5.5.7.1.28"
	newOidASKey        = "1.3.6.1.5.5.7.1.29"
)

//include prefix bytes
var oidManifestKeyByte []byte = []byte{0x06, 0x0B, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x09, 0x10, 0x01, 0x1A}
var oidAkiKeyByte []byte = []byte{0x06, 0x03, 0x55, 0x1D, 0x23}
var oidSkiKeyByte []byte = []byte{0x06, 0x03, 0x55, 0x1D, 0x0E}
var oidCaIssuersKeyByte []byte = []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x05, 0x05, 0x07, 0x30, 0x02}
var oidRpkiManifestKeyByte []byte = []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x05, 0x05, 0x07, 0x30, 0x0A}
var oidRpkiNotifyKeyByte []byte = []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x05, 0x05, 0x07, 0x30, 0x0D}
var oidCaRepositoryKeyByte []byte = []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x05, 0x05, 0x07, 0x30, 0x05}
var oidSignedObjectKeyByte []byte = []byte{0x06, 0x08, 0x2B, 0x06, 0x01, 0x05, 0x05, 0x07, 0x30, 0x0B}

var oidRoaKeyByte []byte = []byte{0x06, 0x0B, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x09, 0x10, 0x01, 0x18}
