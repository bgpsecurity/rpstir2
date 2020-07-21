package rtrtcp

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

func ParseToSerialQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get sessionId fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get SessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 12 {
		belogs.Error("ParseToSerialQuery():PDU_TYPE_SERIAL_QUERY,  length must be 12 ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is SERIAL QUERY, length must be 12"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get serialNumber fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}
	sq := rtrmodel.NewRtrSerialQueryModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialQuery():get PDU_TYPE_SERIAL_QUERY ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}

func ProcessSerialQuery(rtrPduModel rtrmodel.RtrPduModel) (serialResponses []rtrmodel.RtrPduModel, err error) {
	rtrSerialQueryModel, p := rtrPduModel.(*rtrmodel.RtrSerialQueryModel)
	if !p {
		belogs.Error("ProcessSerialQuery(): rtrPduModel convert to rtrResetQueryModel fail ")
		return nil, errors.New("processRtrPduModel(): rtrPduModel convert to rtrResetQueryModel fail  ")
	}
	clientSessionId := rtrSerialQueryModel.SessionId
	clientSerialNumber := rtrSerialQueryModel.SerialNumber
	belogs.Debug("ProcessSerialQuery(): clientSessionId,   clientSerialNum : ", clientSessionId, clientSerialNumber)

	//
	serialNumbers, err := needResetQuery(clientSessionId, clientSerialNumber)
	belogs.Debug("ProcessSerialQuery(): needReset,   clientSessionId, clientSerialNumber,serialNumbers : ", clientSessionId, clientSerialNumber, serialNumbers)
	if err != nil {
		belogs.Error("ProcessSerialQuery(): needResetQuery fail ,  clientSessionId, clientSerialNumber, err:", clientSessionId, clientSerialNumber, err)
		return nil, errors.New("processRtrPduModel(): needResetQuery fail  ")
	}
	belogs.Debug("ProcessSerialQuery():  clientSessionId:", clientSessionId,
		",  clientSerialNumber:", clientSerialNumber, ",server get  serialNumbers : ", serialNumbers)

	//
	if len(serialNumbers) == 0 {
		// no new data, so just send End Of Data PDU
		rtrPduModels := assembleEndOfDataResponses(rtrSerialQueryModel.GetProtocolVersion(), clientSessionId, clientSerialNumber)
		belogs.Info("ProcessSerialQuery(): server get len(serialNumbers) == 0, will just send End Of Data PDU Response,",
			"  clientSessionId: ", clientSessionId, ",  clientSerialNumber:", clientSerialNumber,
			",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
		return rtrPduModels, nil

	} else if len(serialNumbers) > 2 {
		// send Cache Reset PDU Response
		belogs.Debug("ProcessSerialQuery(): server get len(serialNumbers) >2, will send Cache Reset PDU Response,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber, ", len(serialNumbers):", len(serialNumbers))
		rtrPduModels, err := assembleCacheResetResponses(rtrSerialQueryModel.GetProtocolVersion())
		if err != nil {
			belogs.Error("ProcessSerialQuery(): len(serialNumbers) >2, assembleCacheResetResponses , fail: ", err)
			return nil, err
		}
		belogs.Info("ProcessSerialQuery(): server get len(serialNumbers) >2, will send Cache Reset PDU Response,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers):", len(serialNumbers), ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
		return rtrPduModels, nil
	} else if len(serialNumbers) > 0 && len(serialNumbers) <= 2 {
		// send Cache Response
		belogs.Debug("ProcessSerialQuery():server get  len(serialNumbers) >0 && <=2 , will send Cache Response of rtr incremental,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers): ", len(serialNumbers))
		rtrIncrementals, sessionId, serialNumber, err := db.GetRtrIncrementalAndSessionIdAndSerialNumber(clientSerialNumber)
		if err != nil {
			belogs.Error("ProcessSerialQuery(): len(serialNumbers) >0 && <=2,  GetRtrIncrementalAndSessionIdAndSerialNumber fail: ", err)
			return nil, err
		}
		rtrPduModels, err := assembleSerialResponses(rtrIncrementals, rtrSerialQueryModel.GetProtocolVersion(), sessionId, serialNumber)
		if err != nil {
			belogs.Error("ProcessSerialQuery():server get len(serialNumbers) >0 && <=2 , assembleSerialResponses fail: ", err)
			return nil, err
		}
		belogs.Info("ProcessSerialQuery():server get  len(serialNumbers) >0 && <=2 , will send Cache Response of rtr incremental,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers): ", len(serialNumbers), ",  len(rtrPduModels):", len(rtrPduModels))

		return rtrPduModels, nil
	}
	return nil, errors.New("processRtrPduModel(): server get serial number from client is err")
}

// 1: check error;  ;
func needResetQuery(clientSessionId uint16, clientSerialNumber uint32) (serialNumbers []uint32, err error) {

	sessionId, err := db.GetSessionId()
	belogs.Debug("needResetQuery(): sessionId, clientSessionId: ", sessionId, clientSessionId)
	if err != nil {
		belogs.Error("needResetQuery(): GetSessionId fail: ", err)
		return nil, err
	}
	if sessionId != clientSessionId {
		belogs.Debug("judgeRtrIncrAvailable(): sessionId != clientSessionId : ", sessionId, clientSessionId)
		return nil, errors.New("needResetQuery():, sessionId is not equal to clientSessionId")
	}

	serialNumbers, err = db.GetSpanSerialNumbers(clientSerialNumber)
	belogs.Debug("needResetQuery(): GetSpanSerialNumbers clientSerialNumber, serialNumbers : ", clientSerialNumber, serialNumbers)
	if err != nil {
		belogs.Error("needResetQuery(): GetSpanSerialNumbers clientSerialNumber, serialNumbers fail : ", clientSerialNumber, serialNumbers, err)
		return nil, err
	}
	return serialNumbers, nil

}

// when len(rtrIncrementals)==0, just return endofdata, it is not an error
func assembleSerialResponses(rtrIncrementals []model.LabRpkiRtrIncremental, protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModels []rtrmodel.RtrPduModel, err error) {

	belogs.Info("assembleSerialResponses(): len(rtrIncrementals):", len(rtrIncrementals),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	rtrPduModels = make([]rtrmodel.RtrPduModel, 0)

	if len(rtrIncrementals) > 0 {
		belogs.Debug("assembleSerialResponses(): len(rtrIncrementals)>0, len(rtrIncrementals): ", len(rtrIncrementals),
			"  protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)

		cacheResponseModel := rtrmodel.NewRtrCacheResponseModel(protocolVersion, sessionId)
		belogs.Debug("assembleSerialResponses(): cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

		rtrPduModels = append(rtrPduModels, cacheResponseModel)
		for i, _ := range rtrIncrementals {
			rtrPduModel, err := convertRtrIncrementalToRtrPduModel(&rtrIncrementals[i], protocolVersion)
			if err != nil {
				belogs.Error("assembleSerialResponses(): convertRtrFullToRtrPduModel fail: ", err)
				return rtrPduModels, err
			}
			belogs.Debug("assembleSerialResponses(): rtrPduModel : ", jsonutil.MarshalJson(rtrPduModel))

			rtrPduModels = append(rtrPduModels, rtrPduModel)
		}

		endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		belogs.Debug("assembleSerialResponses(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

		rtrPduModels = append(rtrPduModels, endOfDataModel)
		return rtrPduModels, nil
	} else {
		belogs.Debug("assembleSerialResponses(): len(rtrIncrementals)==0 : just send endofdata,",
			" protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
		rtrPduModels = assembleEndOfDataResponses(protocolVersion, sessionId, serialNumber)
		return rtrPduModels, nil
	}

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
