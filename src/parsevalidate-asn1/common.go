package parsevalidateasn1

import (
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/asn1util/asn1base"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type OctetString []byte

// addressPrefix or min/max
type IpAddrBlock struct {
	AddressFamily uint64 `json:"addressFamily" asn1:"optional"`
	//address prefix: 147.28.83.0/24 '
	AddressPrefix string `json:"addressPrefix"`
	//min address:  99.96.0.0
	Min string `json:"min"`
	//max address:   99.105.127.255
	Max string `json:"max"`
}
type IpAddrRaw struct {
	AddressFamily   []byte
	IpAddressChoice asn1.RawValue
}
type IpAddrRange struct {
	Min asn1.BitString
	Max asn1.BitString
}

// ipv4: size==4, ipv6: size==6
func ParseBitStringToIpNet(bi asn1.BitString, ipType int) (ipNet *net.IPNet, err error) {
	var size int
	if ipType == 1 {
		size = 4
	} else if ipType == 2 {
		size = 16
	} else {
		belogs.Error("ParseBitStringToAddressPrefix(): ipType fail:", ipType)
		return nil, errors.New("Not an IP type")
	}

	ipAddr := make([]byte, size)
	copy(ipAddr, bi.Bytes)
	mask := net.CIDRMask(bi.BitLength, size*8)
	belogs.Debug("ParseBitStringToAddressPrefix(): ipAddr:", convert.PrintBytesOneLine(ipAddr),
		jsonutil.MarshalJson(ipAddr), "  mask:", mask)
	return &net.IPNet{
		IP:   net.IP(ipAddr),
		Mask: mask,
	}, nil
}

// data: just ip data, no asn1 header
func ParseBytesToIpNet(data []byte, ipType int) (*net.IPNet, error) {
	belogs.Debug("ParseBytesToIpNet(): ipType:", ipType, "   len(data):", len(data))

	bi, err := asn1base.ParseBitString(data)
	if err != nil {
		belogs.Error("ParseBytesToIpNet(): ParseBitString fail:", convert.PrintBytesOneLine(data))
		return nil, errors.New("data is not IP address")
	}
	bitString := asn1.BitString{
		Bytes:     bi.Bytes,
		BitLength: bi.BitLength,
	}
	return ParseBitStringToIpNet(bitString, ipType)
}

// use ParseBytesToIpNet --> 134.144.0.0/16
func ParseBitStringToAddressPrefix(bi asn1.BitString, ipType int) (addressPrefix string, err error) {
	net, err := ParseBitStringToIpNet(bi, ipType)
	if err != nil {
		belogs.Error("ParseBitStringToAddressPrefix(): ParseBitStringToIpNet fail:", err)
		return "", errors.New("data is not IP address")
	}
	return net.String(), nil
}

// use ParseBytesToIpNet --> 134.144.0.0/16
func ParseBytesToAddressPrefix(data []byte, ipType int) (addressPrefix string, err error) {
	net, err := ParseBytesToIpNet(data, ipType)
	if err != nil {
		belogs.Error("ParseBytesToAddressPrefix(): ParseBytesToIpNet fail:", err)
		return "", errors.New("data is not IP address")
	}
	return net.String(), nil
}

func ParseToAddressMinMax(data []byte, ipType int) (min, max string, err error) {
	belogs.Debug("ParseToAddressMinMax():data:", convert.PrintBytesOneLine(data), "  ipType:", ipType)

	var size int
	if ipType == 1 {
		size = 4
	} else if ipType == 2 {
		size = 16
	} else {
		belogs.Error("ParseToAddressMinMax(): ipType fail:", ipType)
		return "", "", errors.New("Not an IP address")
	}

	var ipAddrRange IpAddrRange
	_, err = asn1.Unmarshal(data, &ipAddrRange)
	if err != nil {
		belogs.Error("ParseToAddressMinMax():Unmarshal ipAddrRange fail:", convert.PrintBytesOneLine(data), err)
		return "", "", errors.New("data is not IP addresses(min/max)")
	}
	belogs.Debug("ParseToAddressMinMax():ipAddrRange:", ipAddrRange, err)

	// get min
	ipAddrMin := make([]byte, size)
	copy(ipAddrMin, ipAddrRange.Min.Bytes)
	netIpMin := net.IP(ipAddrMin)
	belogs.Debug("ParseToAddressMinMax(): netIpMin:", netIpMin.String())

	// get max, and may be set 0xFF
	ipAddrMax := make([]byte, size)
	copy(ipAddrMax, ipAddrRange.Max.Bytes)
	for i := ipAddrRange.Max.BitLength/8 + 1; i < len(ipAddrMax); i++ {
		ipAddrMax[i] = 0xFF
	}
	if ipAddrRange.Max.BitLength/8 > len(ipAddrMax) {
		belogs.Error("ParseToAddressMinMax():max fail, ipAddrRange.Max.BitLength/8 > len(ipAddrMax):", convert.PrintBytesOneLine(ipAddrRange.Max.Bytes),
			"   ipAddrRange.Max.BitLength/8:", ipAddrRange.Max.BitLength/8, " len(ipAddrMax):", len(ipAddrMax))
		return "", "", errors.New("get max fail")
	}
	if ipAddrRange.Max.BitLength/8 < len(ipAddrMax) {
		ipAddrMax[ipAddrRange.Max.BitLength/8] |= 0xFF >> uint(8-(8*(ipAddrRange.Max.BitLength/8+1)-ipAddrRange.Max.BitLength))
	}
	netIpMax := net.IP(ipAddrMax)
	belogs.Debug("ParseToAddressMinMax(): netIpMax:", netIpMax.String())

	return netIpMin.String(), netIpMax.String(), nil

}

type OidAndValueAsn1 struct {
	OidAsn1   asn1.ObjectIdentifier
	ValueAsn1 asn1.RawValue `asn1:"optional"`
}

type SignerInfoAsn1 struct {
	Version                int               `json:"version"`
	Sid                    OctetString       `json:"sid" asn1:"tag:0"`
	DigestAlgorithmAsn1    OidAndValueAsn1   `json:"digestAlgorithm"`
	OidAndValueAsn1s       []OidAndValueAsn1 `json:"attributeTypeAndValues" asn1:"tag:0"`
	SignatureAlgorithmAsn1 OidAndValueAsn1   `json:"signatureAlgorithm"`
	Sinagture              OctetString       `json:"sinagture"`
}

func ParseToSignerInfoModel(data []byte) (signerInfoModel model.SignerInfoModel, err error) {
	// signerInfos SignerInfos
	belogs.Debug("ParseToSignerInfoModel(): data:", convert.PrintBytesOneLine(data))
	var signerInfoAsn1 SignerInfoAsn1
	_, err = asn1.Unmarshal(data, &signerInfoAsn1)
	if err != nil {
		belogs.Error("ParseToSignerInfoModel(): SignerInfoAsn1 fail, len(data):", len(data), err)
		return signerInfoModel, err
	}
	signerInfoModel.Version = signerInfoAsn1.Version
	if signerInfoAsn1.DigestAlgorithmAsn1.OidAsn1.String() == "2.16.840.1.101.3.4.2.1" {
		signerInfoModel.DigestAlgorithm = "sha256"
	}

	belogs.Debug("ParseToSignerInfoModel(): len(OidAndValueAsn1s):", len(signerInfoAsn1.OidAndValueAsn1s))
	for i := range signerInfoAsn1.OidAndValueAsn1s {
		oid := signerInfoAsn1.OidAndValueAsn1s[i].OidAsn1.String()
		value := signerInfoAsn1.OidAndValueAsn1s[i].ValueAsn1
		belogs.Debug("ParseToSignerInfoModel(): rang OidAndValueAsn1s, i:", i, " oid:", oid)
		if oid == "1.2.840.113549.1.9.3" {
			/*
				type AttributeType OBJECT IDENTIFIER 1.2.840.113549.1.9.3 contentType (PKCS #9)
				values SET (1 elem)
				AttributeValue [?] OBJECT IDENTIFIER 1.2.840.113549.1.9.16.1.24 routeOriginAttest (S/MIME Content Types)
			*/
			var valueOid asn1.ObjectIdentifier
			_, err = asn1.Unmarshal(value.Bytes, &valueOid)
			if err != nil {
				belogs.Error("ParseToSignerInfoModel(): Unmarshal valueOid fail, len(value.Bytes):", len(value.Bytes), err)
				continue
			}
			signerInfoModel.ContentType = valueOid.String()
			belogs.Debug("ParseToSignerInfoModel(): signerInfoModel.ContentType:", signerInfoModel.ContentType)

		} else if oid == "1.2.840.113549.1.9.5" {
			var tm time.Time
			_, err = asn1.Unmarshal(value.Bytes, &tm)
			if err != nil {
				belogs.Error("ParseToSignerInfoModel(): Unmarshal tm fail, len(value.Bytes):", len(value.Bytes), err)
				continue
			}
			signerInfoModel.SigningTime = tm
			belogs.Debug("ParseToSignerInfoModel(): signerInfoModel.SigningTime:", signerInfoModel.SigningTime)

		} else if oid == "1.2.840.113549.1.9.4" {
			var oct OctetString
			_, err = asn1.Unmarshal(value.Bytes, &oct)
			if err != nil {
				belogs.Error("ParseToSignerInfoModel(): Unmarshal oct fail, len(value.Bytes):", len(value.Bytes), err)
				continue
			}
			signerInfoModel.MessageDigest = hex.EncodeToString([]byte(oct))
			belogs.Debug("ParseToSignerInfoModel(): signerInfoModel.MessageDigest:", signerInfoModel.MessageDigest)
		}
	}
	belogs.Debug("ParseToSignerInfoModel(): signerInfoModel:", jsonutil.MarshalJson(signerInfoModel))
	return signerInfoModel, nil
}
