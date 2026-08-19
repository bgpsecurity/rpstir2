package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrIpv6PrefixModel struct {
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
}

func NewRtrIpv6PrefixModel(protocolVersion uint8, flags uint8,
	prefixLength uint8, maxLength uint8, ipv6Prefix [16]byte, asn uint32) *RtrIpv6PrefixModel {
	return &RtrIpv6PrefixModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_IPV6_PREFIX,
		Zero0:           0,
		Length:          32,
		Zero1:           0,
		Flags:           flags,
		PrefixLength:    prefixLength,
		MaxLength:       maxLength,
		Ipv6Prefix:      ipv6Prefix,
		Asn:             asn,
	}
}

func (p *RtrIpv6PrefixModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.PrefixLength)
	binary.Write(wr, binary.BigEndian, p.MaxLength)
	binary.Write(wr, binary.BigEndian, p.Zero1)
	binary.Write(wr, binary.BigEndian, p.Ipv6Prefix)
	binary.Write(wr, binary.BigEndian, p.Asn)
	return wr.Bytes()
}
func (p *RtrIpv6PrefixModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrIpv6PrefixModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrIpv6PrefixModel) GetPduType() uint8 {
	return p.PduType
}
func ParseToIpv6Prefix(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
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
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get zero0 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero0")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 32 {
		belogs.Error("ParseToIpv6Prefix():PDU_TYPE_IPV6_PREFIX, length must be 32, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, length must be 32"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get flags fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	if flags != 0 && flags != 1 {
		belogs.Error("ParseToIpv6Prefix():PDU_TYPE_IPV6_PREFIX, flags must be 0 or 1, buf:", buf, "  flags:", flags)
		rtrError := NewRtrError(
			errors.New("pduType is IPV6 PREFIX, flags must be 0 or 1"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get prefixLength
	err = binary.Read(buf, binary.BigEndian, &prefixLength)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get prefixLength fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get prefixLength")
		return rtrPduModel, rtrError
	}

	// get maxLength
	err = binary.Read(buf, binary.BigEndian, &maxLength)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get maxLength fail: ", err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get maxLength")
		return rtrPduModel, rtrError
	}

	// get zero1
	err = binary.Read(buf, binary.BigEndian, &zero1)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get zero1 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero1")
		return rtrPduModel, rtrError
	}

	// get ipv6Prefix
	err = binary.Read(buf, binary.BigEndian, &ipv6Prefix)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get ipv6Prefix fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv6Prefix")
		return rtrPduModel, rtrError
	}

	err = binary.Read(buf, binary.BigEndian, &asn)
	if err != nil {
		belogs.Error("ParseToIpv6Prefix(): PDU_TYPE_IPV6_PREFIX get asn fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asn")
		return rtrPduModel, rtrError
	}

	sq := NewRtrIpv6PrefixModel(protocolVersion, flags, prefixLength,
		maxLength, ipv6Prefix, asn)

	belogs.Debug("ParseToIpv6Prefix():get PDU_TYPE_IPV6_PREFIX, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
