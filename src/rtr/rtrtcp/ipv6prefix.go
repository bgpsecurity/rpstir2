package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

func ParseToIpv6Prefix(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	/*
		ProtocolVersion uint8    `json:"protocolVersion"`
		PduType         uint8    `json:"pduType"`
		Zero0           uint16   `json:"zero0"`
		Length          uint32   `json:"length"`
		Flags           uint8    `json:"flags"`
		PrefixLength    uint8    `json:"prefixLength"`
		MaxLength       uint8    `json:"maxLength"`
		Zero1           uint8    `json:"zero1"`
		Ipv6Prefix      [16]byte `json:"ipv6Prefix"`
		Asn             uint32   `json:"asn"`
	*/
	var zero0 uint16
	var length uint32
	var flags uint8
	var prefixLength uint8
	var maxLength uint8
	var zero1 uint8
	var ipv6Prefix [16]byte
	var asn uint32

	// get zero0
	err = binary.Read(buf, binary.BigEndian, &zero0)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get zero0 fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero0")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 32 {
		belogs.Error("ParseToIpv6Prefix():PDU_TYPE_IPV6_PREFIX, length must be 32,  ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is IPV4 PREFIX, length must be 32"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get flags fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	if flags != 0 && flags != 1 {
		belogs.Error("ParseToIpv6Prefix():PDU_TYPE_IPV6_PREFIX, flags must be 0 or 1, ", buf, flags)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is IPV6 PREFIX, flags must be 0 or 1"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get prefixLength
	err = binary.Read(buf, binary.BigEndian, &prefixLength)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get prefixLength fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get prefixLength")
		return rtrPduModel, rtrError
	}

	// get maxLength
	err = binary.Read(buf, binary.BigEndian, &maxLength)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get maxLength fail: ", err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get maxLength")
		return rtrPduModel, rtrError
	}

	// get zero1
	err = binary.Read(buf, binary.BigEndian, &zero1)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get zero1 fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero1")
		return rtrPduModel, rtrError
	}

	// get ipv6Prefix
	err = binary.Read(buf, binary.BigEndian, &ipv6Prefix)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get ipv6Prefix fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv6Prefix")
		return rtrPduModel, rtrError
	}

	err = binary.Read(buf, binary.BigEndian, &asn)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get asn fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asn")
		return rtrPduModel, rtrError
	}

	sq := rtrmodel.NewRtrIpv6PrefixModel(protocolVersion, flags, prefixLength,
		maxLength, ipv6Prefix, asn)

	belogs.Debug("ParseToIpv6Prefix():get PDU_TYPE_IPV6_PREFIX ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
