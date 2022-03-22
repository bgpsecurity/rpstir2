package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

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
		belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get sessionId fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get sessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get length fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if protocolVersion == PDU_PROTOCOL_VERSION_0 && length != 12 {
		belogs.Error("ParseToEndOfData():PDU_TYPE_END_OF_DATA, when version is 0, length must be 12, ", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is CACHE RESPONSE, when version is 0, length must be 12"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if protocolVersion == PDU_PROTOCOL_VERSION_1 && length != 24 {
		belogs.Error("ParseToEndOfData():PDU_TYPE_END_OF_DATA,   when version is 1, length must be 24, ", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is CACHE RESPONSE, when version is 1, length must be 24"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get serialNumber fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}

	if protocolVersion == PDU_PROTOCOL_VERSION_1 {
		// get refreshInterval
		err = binary.Read(buf, binary.BigEndian, &refreshInterval)
		if err != nil {
			belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get refreshInterval fail: ", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get refreshInterval")
			return rtrPduModel, rtrError
		}

		// get retryInterval
		err = binary.Read(buf, binary.BigEndian, &retryInterval)
		if err != nil {
			belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get retryInterval fail: ", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get retryInterval")
			return rtrPduModel, rtrError
		}

		// get expireInterval
		err = binary.Read(buf, binary.BigEndian, &expireInterval)
		if err != nil {
			belogs.Error("ParseToEndOfData(): PDU_TYPE_END_OF_DATA get expireInterval fail: ", buf, err)
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
	belogs.Debug("ParseToEndOfData():get PDU_TYPE_END_OF_DATA ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
func assembleEndOfDataResponses(protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModels []RtrPduModel) {
	cacheResetResponseModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
	rtrPduModels = make([]RtrPduModel, 0)
	rtrPduModels = append(rtrPduModels, cacheResetResponseModel)
	return rtrPduModels
}

func assembleEndOfDataResponse(protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModel RtrPduModel) {

	endOfDataModel := NewRtrEndOfDataModel(protocolVersion, sessionId,
		serialNumber, PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_RECOMMENDED,
		PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_RECOMMENDED, PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_RECOMMENDED)
	belogs.Debug("assembleEndOfDataResponse(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))
	return endOfDataModel
}
