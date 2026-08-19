package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrCacheResponseModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
}

func NewRtrCacheResponseModel(protocolVersion uint8, sessionId uint16) *RtrCacheResponseModel {
	return &RtrCacheResponseModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_CACHE_RESPONSE,
		SessionId:       sessionId,
		Length:          8,
	}
}

func (p *RtrCacheResponseModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)
	binary.Write(wr, binary.BigEndian, p.Length)
	return wr.Bytes()
}

func (p *RtrCacheResponseModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrCacheResponseModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrCacheResponseModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToCacheResponse(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var sessionId uint16
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToCacheResponse(): PDU_TYPE_CACHE_RESPONSE get sessionId fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get sessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToCacheResponse(): PDU_TYPE_CACHE_RESPONSE get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToCacheResponse():PDU_TYPE_CACHE_RESPONSE,  length must be 8, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is CACHE RESPONSE, length must be 8"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	sq := NewRtrCacheResponseModel(protocolVersion, sessionId)
	belogs.Debug("ParseToCacheResponse():get PDU_TYPE_CACHE_RESPONSE, buf:", buf, " sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
