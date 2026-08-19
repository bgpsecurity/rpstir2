package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrSerialNotifyModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
	SerialNumber    uint32 `json:"serialNumber"`
}

func NewRtrSerialNotifyModel(protocolVersion uint8, sessionId uint16, serialNumber uint32) *RtrSerialNotifyModel {
	return &RtrSerialNotifyModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_SERIAL_NOTIFY,
		SessionId:       sessionId,
		Length:          12,
		SerialNumber:    serialNumber,
	}
}

func (p *RtrSerialNotifyModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SerialNumber)
	return wr.Bytes()
}
func (p *RtrSerialNotifyModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrSerialNotifyModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}
func (p *RtrSerialNotifyModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToSerialNotify(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialNotify(): PDU_TYPE_SERIAL_NOTIFY get sessionId fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get SessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialNotify(): PDU_TYPE_SERIAL_NOTIFY get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 12 {
		belogs.Error("ParseToSerialNotify():PDU_TYPE_SERIAL_NOTIFY,  length must be 12, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is SERIAL NOTIFY,  length must be 12"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialNotify(): PDU_TYPE_SERIAL_NOTIFY get serialNumber fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}

	sq := NewRtrSerialNotifyModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialNotify():get PDU_TYPE_SERIAL_NOTIFY,  buf:", buf, " sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
