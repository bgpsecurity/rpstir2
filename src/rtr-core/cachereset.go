package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrCacheResetModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	Zero            uint16 `json:"zero"`
	Length          uint32 `json:"length"`
}

func NewRtrCacheResetModel(protocolVersion uint8) *RtrCacheResetModel {
	return &RtrCacheResetModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_CACHE_RESET,
		Zero:            0,
		Length:          8,
	}
}

func (p *RtrCacheResetModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	return wr.Bytes()
}

func (p *RtrCacheResetModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrCacheResetModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrCacheResetModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToCacheReset(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var zero uint16
	var length uint32

	// get zero
	err = binary.Read(buf, binary.BigEndian, &zero)
	if err != nil {
		belogs.Error("ParseToCacheReset(): PDU_TYPE_CACHE_RESET get zero fail,  buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToCacheReset(): PDU_TYPE_CACHE_RESET get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToCacheReset():PDU_TYPE_CACHE_RESET,  length must be 8, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is CACHE RESET, length must be 8"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}
	sq := NewRtrCacheResetModel(protocolVersion)
	belogs.Debug("ParseToCacheReset():get PDU_TYPE_CACHE_RESET, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
func AssembleCacheResetResponses(protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	cacheResetResponseModel := NewRtrCacheResetModel(protocolVersion)
	belogs.Debug("AssembleCacheResetResponses(): cacheResetResponseModel : ", jsonutil.MarshalJson(cacheResetResponseModel))
	rtrPduModels = append(rtrPduModels, cacheResetResponseModel)
	return rtrPduModels, nil

}
