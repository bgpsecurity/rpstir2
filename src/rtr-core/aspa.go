package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrAsaModel struct {
	ProtocolVersion uint8    `json:"protocolVersion"`
	PduType         uint8    `json:"pduType"`
	Flags           uint8    `json:"flags"`
	Zero0           uint8    `json:"zero0"`
	Length          uint32   `json:"length"`
	CustomerAsn     uint32   `json:"customerAsn"`
	ProviderAsns    []uint32 `json:"providerAsns"`
}

func NewRtrAsaModel(protocolVersion uint8, flags uint8,
	customerAsn uint32, providerAsns []uint32) *RtrAsaModel {
	length := 12 + 4*len(providerAsns) // header+flags+length+CustomerAsn + 4 * len(providerAsns)

	return &RtrAsaModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_ASA,
		Zero0:           0,
		Length:          uint32(length),
		Flags:           flags,
		CustomerAsn:     customerAsn,
		ProviderAsns:    providerAsns,
	}
}

func (p *RtrAsaModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.CustomerAsn)
	if len(p.ProviderAsns) > 0 {
		for i := range p.ProviderAsns {
			binary.Write(wr, binary.BigEndian, p.ProviderAsns[i])
		}
	}
	return wr.Bytes()
}
func (p *RtrAsaModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrAsaModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrAsaModel) GetPduType() uint8 {
	return p.PduType
}

func (p *RtrAsaModel) GetKey() string {
	return convert.ToString(p.Flags) + "_" + convert.ToString(p.CustomerAsn) + "_" + jsonutil.MarshalJson(p.ProviderAsns)
}

func ParseToAsa(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion uint8    `json:"protocolVersion"`
		PduType         uint8    `json:"pduType"`
		Flags           uint8    `json:"flags"`
		Zero0           uint8   `json:"zero0"`
		Length          uint32   `json:"length"`
		CustomerAsn     uint32   `json:"customerAsn"`
		ProviderAsns    []uint32 `json:"providerAsns"`
	*/

	var flags uint8
	var zero0 uint8
	var length uint32
	var customerAsn uint32
	var providerAsns []uint32

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
		Bit 0: Announce=1/Withdraw=0;
		Bits 1-7: Reserved
	*/
	if flags != 0 && flags != 1 {
		belogs.Error("ParseToAsa():PDU_TYPE_ASA, flags is only use bits, buf:", buf, "  flags:", flags)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, flags is only use bits"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

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
	if length < 12 {
		belogs.Error("ParseToAsa():PDU_TYPE_ASA, length must be more than 16, buf:", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is ASA, length must be more than 16"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
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

	// 在读取 customerAsn 之后
	remaining := int(length) - 12 // 12 = 固定头部(8) + CustomerAsn(4)
	if remaining%4 != 0 {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get providerAsns count,remaining%4 != 0 fail, buf:", buf)
		// 协议错误，长度必须为 4 的倍数
		rtrError := NewRtrError(err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get providerAsns count")
		return rtrPduModel, rtrError
	}
	count := remaining / 4
	providerAsns = make([]uint32, count)
	if count > 0 {
		err = binary.Read(buf, binary.BigEndian, &providerAsns)
		if err != nil {
			belogs.Error("ParseToAsa(): PDU_TYPE_ASA get providerAsn fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get providerAsn")
			return rtrPduModel, rtrError
		}
	}
	// count ==0  is ok, just continue

	sq := NewRtrAsaModel(protocolVersion, flags, customerAsn, providerAsns)

	belogs.Debug("ParseToAsa():get PDU_TYPE_ASA, buf:", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
