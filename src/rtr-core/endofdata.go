package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrEndOfDataModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
	SerialNumber    uint32 `json:"serialNumber"`
	RefreshInterval uint32 `json:"refreshInterval"`
	RetryInterval   uint32 `json:"retryInterval"`
	ExpireInterval  uint32 `json:"expireInterval"`
}

func NewRtrEndOfDataModel(protocolVersion uint8, sessionId uint16,
	serialNumber uint32, refreshInterval uint32,
	retryInterval uint32, expireInterval uint32) *RtrEndOfDataModel {
	if protocolVersion == PDU_PROTOCOL_VERSION_0 {
		return &RtrEndOfDataModel{
			ProtocolVersion: protocolVersion,
			PduType:         PDU_TYPE_END_OF_DATA,
			SessionId:       sessionId,
			Length:          12,
			SerialNumber:    serialNumber,
		}

	} else if protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2 {
		return &RtrEndOfDataModel{
			ProtocolVersion: protocolVersion,
			PduType:         PDU_TYPE_END_OF_DATA,
			SessionId:       sessionId,
			Length:          24,
			SerialNumber:    serialNumber,
			RefreshInterval: refreshInterval,
			RetryInterval:   retryInterval,
			ExpireInterval:  expireInterval,
		}
	}
	return nil

}
func (p *RtrEndOfDataModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)

	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SerialNumber)
	if p.ProtocolVersion == PDU_PROTOCOL_VERSION_1 || p.ProtocolVersion == PDU_PROTOCOL_VERSION_2 {
		binary.Write(wr, binary.BigEndian, p.RefreshInterval)
		binary.Write(wr, binary.BigEndian, p.RetryInterval)
		binary.Write(wr, binary.BigEndian, p.ExpireInterval)
	}

	return wr.Bytes()
}

func (p *RtrEndOfDataModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrEndOfDataModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrEndOfDataModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToEndOfData(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		if protocolVersion == PDU_PROTOCOL_VERSION_0 {
			return &RtrEndOfDataModel{
				ProtocolVersion: protocolVersion,
				PduType:         PDU_TYPE_END_OF_DATA,
				SessionId:       sessionId,
				Length:          12,
				SerialNumber:    serialNumber,
			}

		} else if protocolVersion == PDU_PROTOCOL_VERSION_1 {
			return &RtrEndOfDataModel{
				ProtocolVersion: protocolVersion,
				PduType:         PDU_TYPE_END_OF_DATA,
				SessionId:       sessionId,
				Length:          24,
				SerialNumber:    serialNumber,
				RefreshInterval: refreshInterval,
				RetryInterval:   retryInterval,
				ExpireInterval:  expireInterval,
			}
		}
	*/

	var sessionId uint16
	var length uint32
	var serialNumber uint32
	var refreshInterval uint32
	var retryInterval uint32
	var expireInterval uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get sessionId fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get sessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if protocolVersion == PDU_PROTOCOL_VERSION_0 && length != 12 {
		belogs.Error("ParseToEndOfData():PDU_TYPE_END_OF_DATA, when version is 0, length must be 12, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is END OF DATA, when version is 0, length must be 12"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if (protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2) && length != 24 {
		belogs.Error("ParseToEndOfData():PDU_TYPE_END_OF_DATA,   when version is 1, length must be 24, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is END OF DATA, when version is 1, length must be 24"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get serialNumber fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}

	if protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2 {
		// get refreshInterval
		err = binary.Read(buf, binary.BigEndian, &refreshInterval)
		if err != nil {
			belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get refreshInterval fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get refreshInterval")
			return rtrPduModel, rtrError
		}

		// get retryInterval
		err = binary.Read(buf, binary.BigEndian, &retryInterval)
		if err != nil {
			belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get retryInterval fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get retryInterval")
			return rtrPduModel, rtrError
		}

		// get expireInterval
		err = binary.Read(buf, binary.BigEndian, &expireInterval)
		if err != nil {
			belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get expireInterval fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get expireInterval")
			return rtrPduModel, rtrError
		}
	}

	sq := NewRtrEndOfDataModel(protocolVersion, sessionId,
		serialNumber, refreshInterval,
		retryInterval, expireInterval)
	belogs.Debug("ParseToEndOfData():get PDU_TYPE_END_OF_DATA, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
func AssembleEndOfDataResponses(protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModels []RtrPduModel) {
	cacheResetResponseModel := AssembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
	rtrPduModels = make([]RtrPduModel, 0)
	rtrPduModels = append(rtrPduModels, cacheResetResponseModel)
	return rtrPduModels
}

func AssembleEndOfDataResponse(protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModel RtrPduModel) {

	refreshInterval := uint32(conf.Int("rtr::refreshInterval"))
	if refreshInterval <= 0 {
		refreshInterval = PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_RECOMMENDED
	}
	retryInterval := uint32(conf.Int("rtr::retryInterval"))
	if retryInterval <= 0 {
		retryInterval = PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_RECOMMENDED
	}
	expireInterval := uint32(conf.Int("rtr::expireInterval"))
	if expireInterval <= 0 {
		expireInterval = PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_RECOMMENDED
	}
	belogs.Debug("AssembleEndOfDataResponse(): refreshInterval:", refreshInterval,
		"  retryInterval:", retryInterval, "   expireInterval:", expireInterval)
	endOfDataModel := NewRtrEndOfDataModel(protocolVersion, sessionId,
		serialNumber,
		refreshInterval,
		retryInterval,
		expireInterval)
	belogs.Debug("AssembleEndOfDataResponse(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))
	return endOfDataModel
}
