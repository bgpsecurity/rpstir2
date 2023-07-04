package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func ParseToAsa(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion uint8    `json:"protocolVersion"`
		PduType         uint8    `json:"pduType"`
		Zero0           uint16   `json:"zero0"`
		Length          uint32   `json:"length"`
		Flags           uint8    `json:"flags"`
		Zero1           uint8    `json:"zero1"`
		ProviderAsCount uint16   `json:"providerAsCount"`
		CustomerAsn     uint32   `json:"customerAsn"`
		ProviderAsns    []uint32 `json:"providerAsns"`
	*/

	var zero0 uint16
	var length uint32
	var flags uint8
	var afiFlags uint8
	var providerAsCount uint16
	var customerAsn uint32
	var providerAsns []uint32
	providerAsns = make([]uint32, 0)

	// get zero0
	err = binary.Read(buf, binary.BigEndian, &zero0)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get zero0 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero0")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length < 16 {
		belogs.Error("ParseToAsa():PDU_TYPE_ASA, length must be more than 16, buf:", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is ASA, length must be more than 16"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get flags fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	/*
		Bit     Bit Name
		----    -------------------
		0      AFI (IPv4 == 0, IPv6 == 1)
		1      Announce == 1, Delete == 0
		2-7    Reserved, must be zero
	*/
	if flags != 0 && flags != 1 && flags != 2 && flags != 3 {
		belogs.Error("ParseToAsa():PDU_TYPE_ASA, flags is only use bits, buf:", buf, "  flags:", flags)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, flags is only use bits"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get afiFlags
	err = binary.Read(buf, binary.BigEndian, &afiFlags)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get afiFlags fail:  buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero1")
		return rtrPduModel, rtrError
	}

	// get providerAsCount
	err = binary.Read(buf, binary.BigEndian, &providerAsCount)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get providerAsCount fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get providerAsCount")
		return rtrPduModel, rtrError
	}

	// get customerAsn
	err = binary.Read(buf, binary.BigEndian, &customerAsn)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get customerAsn fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get customerAsn")
		return rtrPduModel, rtrError
	}
	providerAsns = make([]uint32, 0)
	for i := uint16(0); i < providerAsCount; i++ {
		var providerAsn uint32
		err = binary.Read(buf, binary.BigEndian, &providerAsn)
		if err != nil {
			belogs.Error("ParseToAsa(): PDU_TYPE_ASA get providerAsn fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get providerAsn")
			return rtrPduModel, rtrError
		}
		providerAsns = append(providerAsns, providerAsn)
	}
	sq := NewRtrAsaModelFromParse(protocolVersion, flags, afiFlags,
		customerAsn, providerAsns)

	belogs.Debug("ParseToAsa():get PDU_TYPE_ASA, buf:", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
