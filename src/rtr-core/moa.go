package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrMoaModel struct {
	ProtocolVersion         uint8        `json:"protocolVersion"`
	PduType                 uint8        `json:"pduType"`
	Zero0                   uint16       `json:"zero0"`
	Length                  uint32       `json:"length"`
	Flags                   uint8        `json:"flags"`
	IPv6MappingPrefixLength uint8        `json:"ipv6MappingPrefixLength"`
	IPv4PrefixCount         uint8        `json:"ipv4PrefixCount"`
	Zero1                   uint8        `json:"zero1"`
	IPv6Mapping             [16]byte     `json:"ipv6Prefix"`
	IPv4Prefixes            []IPv4Prefix `json:"ipv4Prefixes"`
}

type IPv4Prefix struct {
	IPv4PrefixLength uint8   `json:"prefixLength"`
	Zero2            uint8   `json:"zero2"`
	Zero3            uint16  `json:"zero3"`
	IPv4Prefix       [4]byte `json:"ipv4Prefix"`
}

func NewRtrMoaModelFromDb(protocolVersion uint8, flags uint8,
	ipv6MappingPrefixLength uint8, ipv6Mapping [16]byte) *RtrMoaModel {
	// Fixed header: 1+1+2+4 + 1+1+1+1 + 16 = 28 bytes
	length := 28
	return &RtrMoaModel{
		ProtocolVersion:         protocolVersion,
		PduType:                 PDU_TYPE_MOA,
		Zero0:                   0,
		Length:                  uint32(length),
		Flags:                   flags,
		IPv6MappingPrefixLength: ipv6MappingPrefixLength,
		IPv4PrefixCount:         0,
		Zero1:                   0,
		IPv6Mapping:             ipv6Mapping,
		IPv4Prefixes:            make([]IPv4Prefix, 0),
	}
}

func NewRtrMoaModelFromParse(protocolVersion uint8, flags uint8,
	ipv6MappingPrefixLength uint8, ipv6Mapping [16]byte, ipv4Prefixes []IPv4Prefix) *RtrMoaModel {
	// Fixed: 28 bytes + each IPv4 prefix entry: 1+1+2+4 = 8 bytes
	// Note: RFC draft says 28 + 5*N, but field definitions sum to 8 bytes per entry
	length := 28 + len(ipv4Prefixes)*8

	return &RtrMoaModel{
		ProtocolVersion:         protocolVersion,
		PduType:                 PDU_TYPE_MOA,
		Zero0:                   0,
		Length:                  uint32(length),
		Flags:                   flags,
		IPv6MappingPrefixLength: ipv6MappingPrefixLength,
		IPv4PrefixCount:         uint8(len(ipv4Prefixes)),
		Zero1:                   0,
		IPv6Mapping:             ipv6Mapping,
		IPv4Prefixes:            ipv4Prefixes,
	}
}

func (p *RtrMoaModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.IPv6MappingPrefixLength)
	binary.Write(wr, binary.BigEndian, p.IPv4PrefixCount)
	binary.Write(wr, binary.BigEndian, p.Zero1)
	binary.Write(wr, binary.BigEndian, p.IPv6Mapping)
	if len(p.IPv4Prefixes) > 0 {
		for i := range p.IPv4Prefixes {
			binary.Write(wr, binary.BigEndian, p.IPv4Prefixes[i].IPv4PrefixLength)
			binary.Write(wr, binary.BigEndian, p.IPv4Prefixes[i].Zero2)
			binary.Write(wr, binary.BigEndian, p.IPv4Prefixes[i].Zero3)
			binary.Write(wr, binary.BigEndian, p.IPv4Prefixes[i].IPv4Prefix)
		}
	}
	return wr.Bytes()
}

func (p *RtrMoaModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}

func (p *RtrMoaModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrMoaModel) GetPduType() uint8 {
	return p.PduType
}
func (p *RtrMoaModel) AddIPv4Prefix(ipv4Prefix IPv4Prefix) {
	if p.IPv4Prefixes == nil {
		p.IPv4Prefixes = make([]IPv4Prefix, 0)
	}
	p.IPv4Prefixes = append(p.IPv4Prefixes, ipv4Prefix)
	p.Length += 8 // 1(Length) + 1(Zero2) + 2(Zero3) + 4(Prefix) = 8
	p.IPv4PrefixCount++
	belogs.Debug("AddIPv4Prefix(): ipv4Prefix:", ipv4Prefix, "  length:", p.Length,
		"  len(ipv4Prefixes):", len(p.IPv4Prefixes), "   IPv4PrefixCount:", p.IPv4PrefixCount)
}

func (p *RtrMoaModel) GetKey() string {
	return convert.ToString(p.Flags) + "_" +
		convert.ToString(p.IPv6MappingPrefixLength) + "_" +
		convert.ToString(p.IPv6Mapping)
}
func ParseToMoa(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion         uint8         `json:"protocolVersion"`
		PduType                 uint8         `json:"pduType"`
		Zero0                   uint16        `json:"zero0"`
		Length                  uint32        `json:"length"`
		Flags                   uint8         `json:"flags"`
		IPv6MappingPrefixLength uint8         `json:"ipv6MappingPrefixLength"`
		IPv4PrefixCount         uint8         `json:"ipv4PrefixCount"`
		Zero1                   uint8         `json:"zero1"`
		IPv6Mapping             [16]byte      `json:"ipv6MappingPrefix"`
		IPv4Prefixes            []IPv4Prefix  `json:"ipv4Prefixes"`
	*/

	var zero0 uint16
	var length uint32
	var flags uint8
	var ipv6MappingPrefixLength uint8
	var ipv4PrefixCount uint8
	var zero1 uint8
	var ipv6Mapping [16]byte
	var ipv4Prefixes []IPv4Prefix

	// get zero0
	err = binary.Read(buf, binary.BigEndian, &zero0)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get zero0 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero0")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	// Minimum: 28 bytes (fixed header with no IPv4 prefixes)
	// With N prefixes: 28 + 8*N
	if length < 28 {
		belogs.Error("ParseToMoa():PDU_TYPE_MOA, length must be >= 28, buf:", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is MOA, length must be >= 28"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get flags fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	/*
		Bit     Bit Name
		----    -------------------
		0       Announce == 1, Withdrawal == 0
		1-7     Reserved, must be zero
	*/
	if flags != 0 && flags != 1 {
		belogs.Error("ParseToMoa():PDU_TYPE_MOA, flags is only use bit 0, buf:", buf, "  flags:", flags)
		rtrError := NewRtrError(
			errors.New("pduType is MOA, flags is only use bit 0"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get ipv6MappingPrefixLength
	err = binary.Read(buf, binary.BigEndian, &ipv6MappingPrefixLength)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv6MappingPrefixLength fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv6MappingPrefixLength")
		return rtrPduModel, rtrError
	}
	if ipv6MappingPrefixLength > 128 {
		belogs.Error("ParseToMoa():PDU_TYPE_MOA, ipv6MappingPrefixLength must be <= 128, buf:", buf, "  ipv6MappingPrefixLength:", ipv6MappingPrefixLength)
		rtrError := NewRtrError(
			errors.New("pduType is MOA, ipv6MappingPrefixLength must be <= 128"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv6MappingPrefixLength")
		return rtrPduModel, rtrError
	}

	// get ipv4PrefixCount
	err = binary.Read(buf, binary.BigEndian, &ipv4PrefixCount)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv4PrefixCount fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv4PrefixCount")
		return rtrPduModel, rtrError
	}

	if length != uint32(28+8*ipv4PrefixCount) {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA length != uint32(28+8*ipv4PrefixCount) fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length or ipv4PrefixCount")
		return rtrPduModel, rtrError
	}

	// get zero1
	err = binary.Read(buf, binary.BigEndian, &zero1)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get zero1 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero1")
		return rtrPduModel, rtrError
	}

	// get ipv6Mapping
	err = binary.Read(buf, binary.BigEndian, &ipv6Mapping)
	if err != nil {
		belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv6Mapping fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get ipv6Mapping")
		return rtrPduModel, rtrError
	}

	// get ipv4Prefixes
	ipv4Prefixes = make([]IPv4Prefix, 0)
	for i := uint8(0); i < ipv4PrefixCount; i++ {
		var ipv4Prefix IPv4Prefix

		err = binary.Read(buf, binary.BigEndian, &ipv4Prefix.IPv4PrefixLength)
		if err != nil {
			belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv4Prefix.IPv4PrefixLength fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get ipv4Prefix.IPv4PrefixLength")
			return rtrPduModel, rtrError
		}
		if ipv4Prefix.IPv4PrefixLength > 32 {
			belogs.Error("ParseToMoa():PDU_TYPE_MOA, ipv4Prefix.IPv4PrefixLength must be <= 32, buf:", buf, "  ipv4PrefixLength:", ipv4Prefix.IPv4PrefixLength)
			rtrError := NewRtrError(
				errors.New("pduType is MOA, ipv4Prefix.IPv4PrefixLength must be <= 32"),
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get ipv4Prefix.IPv4PrefixLength")
			return rtrPduModel, rtrError
		}

		err = binary.Read(buf, binary.BigEndian, &ipv4Prefix.Zero2)
		if err != nil {
			belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv4Prefix.Zero2 fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get ipv4Prefix.Zero2")
			return rtrPduModel, rtrError
		}

		err = binary.Read(buf, binary.BigEndian, &ipv4Prefix.Zero3)
		if err != nil {
			belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv4Prefix.Zero3 fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get ipv4Prefix.Zero3")
			return rtrPduModel, rtrError
		}

		err = binary.Read(buf, binary.BigEndian, &ipv4Prefix.IPv4Prefix)
		if err != nil {
			belogs.Error("ParseToMoa(): PDU_TYPE_MOA get ipv4Prefix.IPv4Prefix fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get ipv4Prefix.IPv4Prefix")
			return rtrPduModel, rtrError
		}

		ipv4Prefixes = append(ipv4Prefixes, ipv4Prefix)
	}

	sq := NewRtrMoaModelFromParse(protocolVersion, flags, ipv6MappingPrefixLength,
		ipv6Mapping, ipv4Prefixes)

	belogs.Debug("ParseToMoa():get PDU_TYPE_MOA, buf:", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
