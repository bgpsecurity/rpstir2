package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrIpv4PrefixModel struct {
	ProtocolVersion uint8   `json:"protocolVersion"`
	PduType         uint8   `json:"pduType"`
	Zero0           uint16  `json:"zero0"`
	Length          uint32  `json:"length"`
	Flags           uint8   `json:"flags"`
	PrefixLength    uint8   `json:"prefixLength"`
	MaxLength       uint8   `json:"maxLength"`
	Zero1           uint8   `json:"zero1"`
	Ipv4Prefix      [4]byte `json:"ipv4Prefix"`
	Asn             uint32  `json:"asn"`
}

func NewRtrIpv4PrefixModel(protocolVersion uint8, flags uint8,
	prefixLength uint8, maxLength uint8, ipv4Prefix [4]byte, asn uint32) *RtrIpv4PrefixModel {
	return &RtrIpv4PrefixModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_IPV4_PREFIX,
		Zero0:           0,
		Length:          20,
		Flags:           flags,
		PrefixLength:    prefixLength,
		MaxLength:       maxLength,
		Zero1:           0,
		Ipv4Prefix:      ipv4Prefix,
		Asn:             asn,
	}
}

func (p *RtrIpv4PrefixModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.PrefixLength)
	binary.Write(wr, binary.BigEndian, p.MaxLength)
	binary.Write(wr, binary.BigEndian, p.Zero1)
	binary.Write(wr, binary.BigEndian, p.Ipv4Prefix)
	binary.Write(wr, binary.BigEndian, p.Asn)
	return wr.Bytes()
}
func (p *RtrIpv4PrefixModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrIpv4PrefixModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrIpv4PrefixModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToIpv4Prefix(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion uint8   `json:"protocolVersion"`
		PduType         uint8   `json:"pduType"`
		Zero0           uint16  `json:"zero0"`
		Length          uint32  `json:"length"`
		Flags           uint8   `json:"flags"`
		PrefixLength    uint8   `json:"prefixLength"`
		MaxLength       uint8   `json:"maxLength"`
		Zero1           uint8   `json:"zero1"`
		Ipv4Prefix      [4]byte `json:"ipv4Prefix"`
		Asn             uint32  `json:"asn"`
	*/
	var zero0 uint16
	var length uint32
	var flags uint8
	var prefixLength uint8
	var maxLength uint8
	var zero1 uint8
	var ipv4Prefix [4]byte
	var asn uint32

	// get zero0
	err = binary.Read(buf, binary.BigEndian, &zero0)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get zero0 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero0")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 20 {
		belogs.Error("ParseToIpv4Prefix():PDU_TYPE_IPV4_PREFIX, length must be 20, buf:", buf, "   length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, length must be 20"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get flags fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	if flags != 0 && flags != 1 {
		belogs.Error("ParseToIpv4Prefix():PDU_TYPE_IPV4_PREFIX, flags must be 0 or 1, buf:", buf, flags)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, flags must be 0 or 1"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get prefixLength
	err = binary.Read(buf, binary.BigEndian, &prefixLength)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get prefixLength fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get prefixLength")
		return rtrPduModel, rtrError
	}

	// get maxLength
	err = binary.Read(buf, binary.BigEndian, &maxLength)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get maxLength fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get maxLength")
		return rtrPduModel, rtrError
	}

	// get zero1
	err = binary.Read(buf, binary.BigEndian, &zero1)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get zero1 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero1")
		return rtrPduModel, rtrError
	}

	// get ipv4Prefix
	err = binary.Read(buf, binary.BigEndian, &ipv4Prefix)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get ipv4Prefix fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv4Prefix")
		return rtrPduModel, rtrError
	}

	// get asn
	err = binary.Read(buf, binary.BigEndian, &asn)
	if err != nil {
		belogs.Error("ParseToIpv4Prefix(): PDU_TYPE_IPV4_PREFIX get asn fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asn")
		return rtrPduModel, rtrError
	}

	sq := NewRtrIpv4PrefixModel(protocolVersion, flags, prefixLength,
		maxLength, ipv4Prefix, asn)

	belogs.Debug("ParseToIpv4Prefix():get PDU_TYPE_IPV4_PREFIX, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
