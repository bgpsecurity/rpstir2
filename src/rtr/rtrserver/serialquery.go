package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	iputil "github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"

	"model"
	db "rtr/db"
	rtrmodel "rtr/model"
)

func ParseToSerialQuery(buf *bytes.Reader, protocolVersion uint8) (rtrmodel.RtrPduModel, error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	err := binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get sessionId fail: ", err)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA, nil, nil),
			err
	}

	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get length fail: ", err)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA, nil, nil), err
	}
	if length != 12 {
		belogs.Error("ParseToSerialQuery():PDU_TYPE_SERIAL_QUERY,  length must be 12 ", length)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA, nil, nil),
			errors.New("pduType is SERIAL QUERY,  length must be 12")
	}
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get serialNumber fail: ", err)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA, nil, nil), err
	}
	sq := rtrmodel.NewRtrSerialQueryModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialQuery():get PDU_TYPE_SERIAL_QUERY ", jsonutil.MarshalJson(sq))
	return sq, nil
}

func ProcessSerialQuery(rtrPduModel rtrmodel.RtrPduModel) (serialResponses []rtrmodel.RtrPduModel, err error) {
	rtrSerialQueryModel, p := rtrPduModel.(*rtrmodel.RtrSerialQueryModel)
	if !p {
		belogs.Error("ProcessSerialQuery(): rtrPduModel convert to rtrResetQueryModel fail ")
		return serialResponses, errors.New("processRtrPduModel(): rtrPduModel convert to rtrResetQueryModel fail  ")
	}
	clientSessionId := rtrSerialQueryModel.SessionId
	clientSerialNumber := rtrSerialQueryModel.SerialNumber
	belogs.Debug("ProcessSerialQuery(): clientSessionId,   clientSerialNum : ", clientSessionId, clientSerialNumber)

	needReset, clientSerialNumId, err := needResetQuery(clientSessionId, clientSerialNumber)
	belogs.Debug("ProcessSerialQuery(): needReset,   clientSerialNumId : ", needReset, clientSerialNumId)
	if err != nil {
		belogs.Error("ProcessSerialQuery(): rtrPduModel convert to rtrResetQueryModel fail ")
		return serialResponses, errors.New("processRtrPduModel(): rtrPduModel convert to rtrResetQueryModel fail  ")
	}
	if needReset {
		rtrPduModels, err := assembleCacheResetResponses(rtrSerialQueryModel.GetProtocolVersion())
		if err != nil {
			belogs.Error("ProcessSerialQuery(): assembleCacheResetResponses fail: ", err)
			return serialResponses, err
		}
		return rtrPduModels, nil
	}
	rtrIncrementals, sessionId, serialNumber, err := db.GetRtrIncrementalAndSessionIdAndSerialNumber(clientSerialNumId)
	if err != nil {
		belogs.Error("ProcessSerialQuery(): GetRtrIncrementalAndSessionIdAndSerialNumber fail: ", err)
		return serialResponses, err
	}
	rtrPduModels, err := assembleSerialResponses(&rtrIncrementals, rtrSerialQueryModel.GetProtocolVersion(), sessionId, serialNumber)
	if err != nil {
		belogs.Error("ProcessSerialQuery(): assembleSerialResponses fail: ", err)
		return serialResponses, err
	}
	return rtrPduModels, nil
}

// 1: check error;  2. check needReset; 3 use clientSerialNumId
func needResetQuery(clientSessionId uint16, clientSerialNumber uint32) (needReset bool, clientSerialNumId int64, err error) {

	sessionId, err := db.GetSessionId()
	belogs.Debug("needResetQuery(): sessionId, clientSessionId: ", sessionId, clientSessionId)
	if err != nil {
		belogs.Error("judgeRtrIncrAvailable(): GetSessionId fail: ", err)
		return false, -1, err
	}
	if sessionId != clientSessionId {
		belogs.Debug("judgeRtrIncrAvailable(): sessionId != clientSessionId : ", sessionId, clientSessionId)
		return true, -1, nil
	}

	// get clientSerialNum --> id
	clientSerialNumId, err = db.ExistSerialNumber(clientSerialNumber)
	belogs.Debug("needResetQuery(): ExistSerialNumber get clientSerialNumId, clientSerialNumber : ", clientSerialNumId, clientSerialNumber)
	if err != nil {
		return false, -1, err
	}
	if clientSerialNumId < 0 {
		return true, -1, nil
	}

	serialNumbers, err := db.GetSpanSerialNumbers(clientSerialNumId)
	belogs.Debug("needResetQuery(): GetSpanSerialNumbers serialNumbers ,clientSerialNumId : ", serialNumbers, clientSerialNumId)
	if err != nil {
		return false, -1, err
	}
	// max serailnumber span is 2
	if len(serialNumbers) > 2 {
		return true, -1, nil
	}
	return false, clientSerialNumId, nil
}

func assembleCacheResetResponses(protocolVersion uint8) (rtrPduModels []rtrmodel.RtrPduModel, err error) {
	rtrPduModels = make([]rtrmodel.RtrPduModel, 0)
	cacheResetResponseModel := rtrmodel.NewRtrCacheResetModel(protocolVersion)
	belogs.Debug("assembleCacheResetResponses(): cacheResetResponseModel : ", jsonutil.MarshalJson(cacheResetResponseModel))
	rtrPduModels = append(rtrPduModels, cacheResetResponseModel)
	return rtrPduModels, nil

}

func assembleSerialResponses(rtrIncrementals *[]model.LabRpkiRtrIncremental, protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModels []rtrmodel.RtrPduModel, err error) {
	rtrPduModels = make([]rtrmodel.RtrPduModel, 0)

	if len(*rtrIncrementals) > 0 {

		cacheResponseModel := rtrmodel.NewRtrCacheResponseModel(protocolVersion)
		belogs.Debug("assembleSerialResponses(): cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

		rtrPduModels = append(rtrPduModels, cacheResponseModel)
		for _, one := range *rtrIncrementals {
			rtrPduModel, err := convertRtrIncrementalToRtrPduModel(&one, protocolVersion)
			if err != nil {
				belogs.Error("assembleSerialResponses(): convertRtrFullToRtrPduModel fail: ", err)
				return rtrPduModels, err
			}
			belogs.Debug("assembleSerialResponses(): rtrPduModel : ", jsonutil.MarshalJson(rtrPduModel))

			rtrPduModels = append(rtrPduModels, rtrPduModel)
		}

		endOfDataModel := rtrmodel.NewRtrEndOfDataModel(protocolVersion, sessionId,
			serialNumber, rtrmodel.PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_RECOMMENDED,
			rtrmodel.PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_RECOMMENDED, rtrmodel.PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_RECOMMENDED)
		belogs.Debug("assembleSerialResponses(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

		rtrPduModels = append(rtrPduModels, endOfDataModel)
	} else {
		errorReportModel := rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)
		belogs.Debug("assembleSerialResponses(): errorReportModel : ", jsonutil.MarshalJson(errorReportModel))

		rtrPduModels = append(rtrPduModels, errorReportModel)
	}
	return rtrPduModels, nil
}

func convertRtrIncrementalToRtrPduModel(rtrIncremental *model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrIncremental.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := rtrmodel.NewRtrIpv4PrefixModel(protocolVersion, rtrmodel.GetIpPrefixModelFlags(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv4, uint32(rtrIncremental.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := rtrmodel.NewRtrIpv6PrefixModel(protocolVersion, rtrmodel.GetIpPrefixModelFlags(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv6, uint32(rtrIncremental.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}
