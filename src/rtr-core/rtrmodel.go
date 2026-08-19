package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/iputil"
)

const (
	PDU_PROTOCOL_VERSION_0 = 0
	PDU_PROTOCOL_VERSION_1 = 1
	PDU_PROTOCOL_VERSION_2 = 2

	PDU_TYPE_SERIAL_NOTIFY  = 0
	PDU_TYPE_SERIAL_QUERY   = 1
	PDU_TYPE_RESET_QUERY    = 2
	PDU_TYPE_CACHE_RESPONSE = 3
	PDU_TYPE_IPV4_PREFIX    = 4
	PDU_TYPE_IPV6_PREFIX    = 6
	PDU_TYPE_END_OF_DATA    = 7
	PDU_TYPE_CACHE_RESET    = 8
	//PDU_TYPE_RESERVED       = 9
	PDU_TYPE_ROUTER_KEY   = 9
	PDU_TYPE_ERROR_REPORT = 10
	PDU_TYPE_ASA          = 11
	PDU_TYPE_MOA          = 12
	PDU_TYPE_END          = 13 // not realy type, just end of type
	// extend

	// min pdu type length is reset query
	PDU_TYPE_MIN_LEN = 8

	// flag: from style
	PDU_FLAG_WITHDRAW = 0
	PDU_FLAG_ANNOUNCE = 1

	// error code
	PDU_TYPE_ERROR_CODE_CORRUPT_DATA                    = 0
	PDU_TYPE_ERROR_CODE_INTERNAL_ERROR                  = 1
	PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE               = 2
	PDU_TYPE_ERROR_CODE_INVALID_REQUEST                 = 3
	PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION    = 4
	PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE            = 5
	PDU_TYPE_ERROR_CODE_WITHDRAWAL_OF_UNKNOWN_RECORD    = 6
	PDU_TYPE_ERROR_CODE_DUPLICATE_ANNOUNCEMENT_RECEIVED = 7
	PDU_TYPE_ERROR_CODE_UNEXPECTED_PROTOCOL_VERSION     = 8

	// seconds.
	PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_MIN         = 1
	PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_MAX         = 86400
	PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_RECOMMENDED = 3600

	PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_MIN         = 1
	PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_MAX         = 7200
	PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_RECOMMENDED = 600

	PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_MIN         = 600
	PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_MAX         = 172800
	PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_RECOMMENDED = 7200

	UINT32_MAX = ^uint32(0)
)

type RtrPduModel interface {
	Bytes() []byte
	PrintBytes() string
	GetProtocolVersion() uint8
	GetPduType() uint8
}

// withdraw-->0, announce-->1
func getModelFlagsFromStyle(style string) uint8 {
	switch style {
	case "withdraw":
		return PDU_FLAG_WITHDRAW
	case "announce":
		return PDU_FLAG_ANNOUNCE
	}
	return 0
}

// rtrPduModel:
func ParseToRtrPduModel(buf *bytes.Reader) (rtrPduModel RtrPduModel, err error) {

	// get length
	if buf.Size() < PDU_TYPE_MIN_LEN {
		belogs.Error("ParseToRtrPduModel(): recv byte's length is too small: ", buf.Size())
		rtrError := NewRtrError(
			errors.New("length of receive bytes is too small"),
			false, uint8(conf.Int("rtr::protocolVersion")), PDU_TYPE_ERROR_CODE_INVALID_REQUEST,
			buf, "")
		return rtrPduModel, rtrError
	}

	// get protocolVersion, pduType
	protocolVersion, pduType, err := parseProtocolVersionAndPduType(buf)
	if err != nil {
		belogs.Error("ParseToRtrPduModel(): parseProtocolVersionAndPduType fail: ", err)
		return rtrPduModel, err
	}

	belogs.Info("ParseToRtrPduModel(): protocolVersion:", protocolVersion, " pduType:", pduType)
	switch pduType {
	case PDU_TYPE_SERIAL_NOTIFY:
		return ParseToSerialNotify(buf, protocolVersion)

	case PDU_TYPE_SERIAL_QUERY:
		return ParseToSerialQuery(buf, protocolVersion)

	case PDU_TYPE_RESET_QUERY:
		return ParseToResetQuery(buf, protocolVersion)

	case PDU_TYPE_CACHE_RESPONSE:
		return ParseToCacheResponse(buf, protocolVersion)

	case PDU_TYPE_IPV4_PREFIX:
		if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
			protocolVersion == PDU_PROTOCOL_VERSION_1 ||
			protocolVersion == PDU_PROTOCOL_VERSION_2 {
			return ParseToIpv4Prefix(buf, protocolVersion)
		}

	case PDU_TYPE_IPV6_PREFIX:
		if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
			protocolVersion == PDU_PROTOCOL_VERSION_1 ||
			protocolVersion == PDU_PROTOCOL_VERSION_2 {
			return ParseToIpv6Prefix(buf, protocolVersion)
		}

	case PDU_TYPE_END_OF_DATA:
		return ParseToEndOfData(buf, protocolVersion)

	case PDU_TYPE_CACHE_RESET:
		return ParseToCacheReset(buf, protocolVersion)

	case PDU_TYPE_ROUTER_KEY:
		return ParseToRouterKey(buf, protocolVersion)

	case PDU_TYPE_ERROR_REPORT:
		return ParseToErrorReport(buf, protocolVersion)

	case PDU_TYPE_ASA:
		if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
			protocolVersion == PDU_PROTOCOL_VERSION_1 ||
			protocolVersion == PDU_PROTOCOL_VERSION_2 {
			return ParseToAsa(buf, protocolVersion)
		}
	case PDU_TYPE_MOA:
		if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
			protocolVersion == PDU_PROTOCOL_VERSION_1 ||
			protocolVersion == PDU_PROTOCOL_VERSION_2 {
			return ParseToMoa(buf, protocolVersion)
		}
	}

	belogs.Error("parseToRtrPduModel():received bytes cannot be parse to rtr's pdu,  pduType:", pduType)
	rtrError := NewRtrError(
		errors.New("received bytes cannot be parse to rtr's pdu, is "+strconv.Itoa(int(pduType))),
		false, protocolVersion, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
		buf, "Fail to get pdu type")
	return rtrPduModel, rtrError
}

func parseProtocolVersionAndPduType(buf *bytes.Reader) (protocolVersion, pduType uint8, err error) {

	// get protocol version
	err = binary.Read(buf, binary.BigEndian, &protocolVersion)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, uint8(conf.Int("rtr::protocolVersion")), PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get protocolVersion")
		return 0, 0, rtrError
	}
	belogs.Debug("parseToPduModel():Read protocolVersion:", protocolVersion)

	if protocolVersion != PDU_PROTOCOL_VERSION_0 &&
		protocolVersion != PDU_PROTOCOL_VERSION_1 &&
		protocolVersion != PDU_PROTOCOL_VERSION_2 {
		belogs.Error("parseToPduModel(): protocolVersion is illegal, buf:", buf, protocolVersion)
		rtrError := NewRtrError(
			errors.New("protocolVersion is illegal, "+strconv.Itoa(int(protocolVersion))),
			true, uint8(conf.Int("rtr::protocolVersion")), PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get protocolVersion")
		return 0, 0, rtrError
	}

	// get pdu type
	err = binary.Read(buf, binary.BigEndian, &pduType)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	belogs.Debug("parseToPduModel():Read pduType:", pduType)

	if pduType >= PDU_TYPE_END {
		belogs.Error("parseToPduModel(): pduType is illegal, buf:", buf, pduType)
		rtrError := NewRtrError(
			errors.New("get Itoa is error "+strconv.Itoa(int(pduType))),
			true, uint8(conf.Int("rtr::protocolVersion")), PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	/*
		if pduType == PDU_TYPE_ROUTER_KEY && protocolVersion == 0 {
			belogs.Error("parseToPduModel():pduType is PDU_TYPE_ROUTER_KEY,  protocolVersion must be more than 0, buf:", buf,
				"  pduType:", pduType, "  protocolVersion:", protocolVersion)
			rtrError := NewRtrError(
				errors.New("pduType is ROUTER KEY,  protocolVersion must be more than 0"),
				true, uint8(conf.Int("rtr::protocolVersion")), PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
				buf, "Fail to get pduType")
			return 0, 0, rtrError
		}

		if pduType == PDU_TYPE_ASA && (protocolVersion == 0 || protocolVersion == 1) {
			belogs.Error("parseToPduModel():pduType is PDU_TYPE_ASA,  protocolVersion must be more than 1, buf:", buf,
				"  pduType:", pduType, "  protocolVersion:", protocolVersion)
			rtrError := NewRtrError(
				errors.New("pduType is PDU_TYPE_ASA,  protocolVersion must be more than 1"),
				true, uint8(conf.Int("rtr::protocolVersion")), PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
				buf, "Fail to get pduType")
			return 0, 0, rtrError
		}
	*/
	belogs.Debug("parseToPduModel():protocolVersion is ", protocolVersion, "  pduType is ", pduType)
	return protocolVersion, pduType, nil
}

func convertAddressPrefixToRtrFormatBytes(addressPrefix string) (rtrFormatBytes []byte, prefixLength uint64, err error) {
	belogs.Debug("convertAddressPrefixToRtrFormatBytes(): addressPrefix",
		addressPrefix)
	prefixStr, prefixLength, err := iputil.SplitAddressAndPrefix(addressPrefix)
	if err != nil {
		belogs.Error("convertAddressPrefixToRtrFormatBytes(): SplitAddressAndPrefix fail, addressPrefix:",
			addressPrefix, err)
		return nil, 0, err
	}

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(prefixStr)
	if err != nil {
		belogs.Error("convertAddressPrefixToRtrFormatBytes(): AddressToRtrFormatByte fail, prefixStr:",
			prefixStr, err)
		return nil, 0, err
	}
	belogs.Debug("convertAddressPrefixToRtrFormatBytes(): AddressToRtrFormatByte ok, addressPrefix",
		addressPrefix, "ipHex", ipHex, "ipType", ipType)
	return ipHex, prefixLength, nil
	/*
		if ipType == iputil.Ipv4Type {
			ipv4 := [4]byte{0x00}
			copy(ipv4[:], ipHex[:])
		} else {
			ipv6 := [16]byte{0x00}
			copy(ipv6[:], ipHex[:])
		}
	*/

}
