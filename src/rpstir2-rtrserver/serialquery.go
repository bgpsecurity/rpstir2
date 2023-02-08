package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

func ParseToSerialQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get sessionId fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get SessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get length fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 12 {
		belogs.Error("ParseToSerialQuery():PDU_TYPE_SERIAL_QUERY,  length must be 12 ", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is SERIAL QUERY, length must be 12"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get serialNumber fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}
	sq := NewRtrSerialQueryModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialQuery():get PDU_TYPE_SERIAL_QUERY ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}

func ProcessSerialQuery(rtrPduModel RtrPduModel) (serialResponses []RtrPduModel, err error) {
	start := time.Now()
	rtrSerialQueryModel, p := rtrPduModel.(*RtrSerialQueryModel)
	if !p {
		belogs.Error("ProcessSerialQuery(): rtrPduModel convert to rtrResetQueryModel fail ")
		return nil, errors.New("processRtrPduModel(): rtrPduModel convert to rtrResetQueryModel fail  ")
	}
	clientSessionId := rtrSerialQueryModel.SessionId
	clientSerialNumber := rtrSerialQueryModel.SerialNumber
	belogs.Info("ProcessSerialQuery(): clientSessionId:", clientSessionId, "  clientSerialNum:", clientSerialNumber)

	//
	serialNumbers, err := needResetQuery(clientSessionId, clientSerialNumber)
	belogs.Debug("ProcessSerialQuery(): needReset,   clientSessionId, clientSerialNumber,serialNumbers : ", clientSessionId, clientSerialNumber, serialNumbers)
	if err != nil {
		belogs.Error("ProcessSerialQuery(): needResetQuery fail ,  clientSessionId, clientSerialNumber, err:", clientSessionId, clientSerialNumber, err)
		return nil, errors.New("processRtrPduModel(): needResetQuery fail  ")
	}
	belogs.Info("ProcessSerialQuery():  clientSessionId:", clientSessionId, ",  clientSerialNumber:", clientSerialNumber,
		"  server get serialNumbers between client and server: ", jsonutil.MarshalJson(serialNumbers),
		"  time(s):", time.Since(start))

	//
	if len(serialNumbers) == 0 {
		// no new data, so just send End Of Data PDU
		rtrPduModels := assembleEndOfDataResponses(rtrSerialQueryModel.GetProtocolVersion(), clientSessionId, clientSerialNumber)
		belogs.Info("ProcessSerialQuery(): server get len(serialNumbers) == 0, will just send End Of Data PDU Response,",
			"  clientSessionId: ", clientSessionId, ",  clientSerialNumber:", clientSerialNumber,
			",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
		return rtrPduModels, nil

	} else if len(serialNumbers) > 2 {
		// shloud send Cache Reset PDU Response
		belogs.Debug("ProcessSerialQuery(): server get len(serialNumbers) >2, will send Cache Reset PDU Response,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber, ", len(serialNumbers):", len(serialNumbers))
		rtrPduModels, err := assembleCacheResetResponses(rtrSerialQueryModel.GetProtocolVersion())
		if err != nil {
			belogs.Error("ProcessSerialQuery(): len(serialNumbers) >2, assembleCacheResetResponses , fail: ", err)
			return nil, err
		}
		belogs.Info("ProcessSerialQuery(): server get len(serialNumbers) >2, will send Cache Reset PDU Response,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers):", len(serialNumbers), ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
		return rtrPduModels, nil
	} else if len(serialNumbers) > 0 && len(serialNumbers) <= 2 {
		// send Cache Response
		belogs.Debug("ProcessSerialQuery():server get  len(serialNumbers) >0 && <=2 , will send Cache Response of rtr incremental,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers): ", len(serialNumbers))
		rtrIncrementals, rtrAsaIncrementals, sessionId, serialNumber, err := getRtrIncrementalAndSessionIdAndSerialNumberDb(clientSerialNumber)
		if err != nil {
			belogs.Error("ProcessSerialQuery(): len(serialNumbers) >0 && <=2,  getRtrIncrementalAndSessionIdAndSerialNumberDb fail: ", clientSerialNumber, err)
			return nil, err
		}
		belogs.Debug("ProcessSerialQuery(): len(rtrIncrementals):", len(rtrIncrementals),
			"  len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "   sessionId:", sessionId, "  serialNumber:", serialNumber)

		rtrPduModels, err := assembleSerialResponses(rtrIncrementals, rtrAsaIncrementals,
			rtrSerialQueryModel.GetProtocolVersion(), sessionId, serialNumber)
		if err != nil {
			belogs.Error("ProcessSerialQuery():server get len(serialNumbers) >0 && <=2 , assembleSerialResponses fail: ", err)
			return nil, err
		}
		belogs.Info("ProcessSerialQuery():server get  len(serialNumbers) >0 && <=2 , will send Cache Response of rtr incremental,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers): ", len(serialNumbers), ",  len(rtrPduModels):", len(rtrPduModels), "  time(s):", time.Since(start))

		return rtrPduModels, nil
	}
	return nil, errors.New("processRtrPduModel(): server get serial number from client is err")
}

// 1: check error;  ;
func needResetQuery(clientSessionId uint16, clientSerialNumber uint32) (serialNumbers []uint32, err error) {

	sessionId, err := getSessionIdDb()
	belogs.Debug("needResetQuery(): sessionId, clientSessionId: ", sessionId, clientSessionId)
	if err != nil {
		belogs.Error("needResetQuery(): getSessionIdDb fail: ", err)
		return nil, err
	}
	if sessionId != clientSessionId {
		belogs.Debug("judgeRtrIncrAvailable(): sessionId != clientSessionId : ", sessionId, clientSessionId)
		return nil, errors.New("needResetQuery():, sessionId is not equal to clientSessionId")
	}

	serialNumbers, err = getSpanSerialNumbersDb(clientSerialNumber)
	belogs.Debug("needResetQuery(): getSpanSerialNumbersDb clientSerialNumber, serialNumbers : ", clientSerialNumber, serialNumbers)
	if err != nil {
		belogs.Error("needResetQuery(): getSpanSerialNumbersDb clientSerialNumber, serialNumbers fail : ", clientSerialNumber, serialNumbers, err)
		return nil, err
	}
	return serialNumbers, nil

}

// when len(rtrIncrementals)==0, just return endofdata, it is not an error
func assembleSerialResponses(rtrIncrementals []model.LabRpkiRtrIncremental, rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	protocolVersion uint8, sessionId uint16, serialNumber uint32) (rtrPduModels []RtrPduModel, err error) {

	belogs.Info("assembleSerialResponses(): len(rtrIncrementals):", len(rtrIncrementals),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	rtrPduModels = make([]RtrPduModel, 0)

	//rtr incr from roa rtr
	if protocolVersion == PDU_PROTOCOL_VERSION_0 || protocolVersion == PDU_PROTOCOL_VERSION_1 {
		if len(rtrIncrementals) > 0 {
			belogs.Debug("assembleSerialResponses(): protocolVersion=0 or 1, len(rtrIncrementals)>0, len(rtrIncrementals): ", len(rtrIncrementals),
				"  protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// start response
			cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
			rtrPduModels = append(rtrPduModels, cacheResponseModel)
			belogs.Debug("assembleSerialResponses(): protocolVersion=0 or 1, cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

			// rtr incr to response
			rtrIncrementalPduModels, err := convertRtrIncrementalsToRtrPduModels(rtrIncrementals, protocolVersion)
			if err != nil {
				belogs.Error("assembleSerialResponses(): protocolVersion=0 or 1, convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrIncrementalPduModels...)
			belogs.Debug("assembleSerialResponses(): protocolVersion=0 or 1, len(rtrIncrementalPduModels) : ", len(rtrIncrementalPduModels))

			// end response
			endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
			rtrPduModels = append(rtrPduModels, endOfDataModel)
			belogs.Debug("assembleSerialResponses(): protocolVersion=0 or 1, endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

			belogs.Info("assembleSerialResponses(): protocolVersion=0 or 1, will send will send Cache Response of incr rtr,",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber,
				",  len(rtrIncrementals): ", len(rtrIncrementals), ",  len(rtrPduModels):", len(rtrPduModels))
			belogs.Debug("assembleSerialResponses(): protocolVersion=0 or 1,  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil
		} else {
			belogs.Debug("assembleSerialResponses(): protocolVersion=0 or 1,len(rtrIncrementals)==0 : just send endofdata,",
				" protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
			rtrPduModels = assembleEndOfDataResponses(protocolVersion, sessionId, serialNumber)

			belogs.Info("assembleSerialResponses(): protocolVersion=0 or 1,there is no rtr this time,  will send errorReport with not_data_available, ",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil
		}
	} else if protocolVersion == PDU_PROTOCOL_VERSION_2 {
		//rtr incr from asa rtr
		if len(rtrIncrementals) > 0 || len(rtrAsaIncrementals) > 0 {
			belogs.Debug("assembleSerialResponses(): protocolVersion=2, len(rtrIncrementals)>0, len(rtrIncrementals): ", len(rtrIncrementals),
				"  protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// start response
			cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
			rtrPduModels = append(rtrPduModels, cacheResponseModel)
			belogs.Debug("assembleSerialResponses(): cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

			// from rtr incr
			rtrIncrementalPduModels, err := convertRtrIncrementalsToRtrPduModels(rtrIncrementals, protocolVersion)
			if err != nil {
				belogs.Error("assembleSerialResponses(): convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrIncrementalPduModels...)
			belogs.Debug("assembleSerialResponses(): len(rtrIncrementalPduModels) : ", len(rtrIncrementalPduModels))

			// rtr asa incr to response
			rtrAsaIncrementalPduModels, err := convertRtrAsaIncrementalsToRtrPduModels(rtrAsaIncrementals, protocolVersion)
			if err != nil {
				belogs.Error("assembleSerialResponses(): convertRtrAsaIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsaIncrementalPduModels...)
			belogs.Debug("assembleSerialResponses(): len(rtrAsaIncrementalPduModels) : ", len(rtrAsaIncrementalPduModels))

			// end response
			endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
			rtrPduModels = append(rtrPduModels, endOfDataModel)
			belogs.Debug("assembleSerialResponses(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

			belogs.Info("assembleResetResponses(): protocolVersion=2, will send will send Cache Response of all rtr,",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber,
				",  len(rtrIncrementals): ", len(rtrIncrementals), ", len(rtrAsaIncrementals): ", len(rtrAsaIncrementals),
				",  len(rtrPduModels):", len(rtrPduModels))
			belogs.Debug("assembleSerialResponses(): protocolVersion=2,  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil
		} else {
			belogs.Debug("assembleSerialResponses(): protocolVersion=2, len(rtrIncrementals)==0 : just send endofdata,",
				" protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
			rtrPduModels = assembleEndOfDataResponses(protocolVersion, sessionId, serialNumber)

			belogs.Info("assembleSerialResponses(): protocolVersion=2,there is no rtr this time,  will send errorReport with not_data_available, ",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil
		}
	}

	belogs.Error("assembleSerialResponses(): not support protocolVersion, fail: ", protocolVersion)
	return nil, errors.New("protocolVersion is not support")

}

func convertRtrIncrementalsToRtrPduModels(rtrIncrementals []model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i, _ := range rtrIncrementals {
		rtrPduModel, err := convertRtrIncrementalToRtrPduModel(&rtrIncrementals[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrIncrementalsToRtrPduModels(): convertRtrIncrementalToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrIncrementalsToRtrPduModels(): len(rtrIncrementals): ", len(rtrIncrementals), " len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrIncrementalToRtrPduModel(rtrIncremental *model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrIncremental.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := NewRtrIpv4PrefixModel(protocolVersion, getModelFlagsFromStyle(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv4, uint32(rtrIncremental.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := NewRtrIpv6PrefixModel(protocolVersion, getModelFlagsFromStyle(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv6, uint32(rtrIncremental.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}

func convertRtrAsaIncrementalsToRtrPduModels(rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	protocolVersion uint8) (rtrAsaPduModels []RtrPduModel, err error) {
	rtrAsaPduModels = make([]RtrPduModel, 0)
	for i, _ := range rtrAsaIncrementals {
		rtrPduModel, err := convertRtrAsaIncrementalToRtrPduModel(&rtrAsaIncrementals[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrIncrementalsToRtrPduModels(): convertRtrIncrementalToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrAsaPduModels = append(rtrAsaPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrIncrementalsToRtrPduModels(): len(rtrAsaIncrementals): ", len(rtrAsaIncrementals),
		" len(rtrAsaPduModels):", len(rtrAsaPduModels))
	return rtrAsaPduModels, nil
}

func convertRtrAsaIncrementalToRtrPduModel(rtrAsaIncremental *model.LabRpkiRtrAsaIncremental,
	protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	providerAsnUint32s := make([]uint32, 0)
	providerAsns := make([]model.ProviderAsn, 0)
	jsonutil.UnmarshalJson(rtrAsaIncremental.ProviderAsns, &providerAsns)
	for i := range providerAsns {
		p := uint32(providerAsns[i].ProviderAsn)
		providerAsnUint32s = append(providerAsnUint32s, p)
	}
	rtrAsaModel := NewRtrAsaModel(protocolVersion, getModelFlagsFromStyle(rtrAsaIncremental.Style),
		uint32(rtrAsaIncremental.CustomerAsn), providerAsnUint32s)
	return rtrAsaModel, nil
}
