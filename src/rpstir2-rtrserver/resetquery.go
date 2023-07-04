package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
)

func ParseToResetQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var zero16 uint16
	var length uint32

	// get zero16
	err = binary.Read(buf, binary.BigEndian, &zero16)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get zero fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToResetQuery():PDU_TYPE_RESET_QUERY, length must be 8, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is RESET QUERY, length must be 8"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	rq := NewRtrResetQueryModel(protocolVersion)
	belogs.Debug("ParseToResetQuery():get PDU_TYPE_RESET_QUERY, buf:", buf, "   rq:", jsonutil.MarshalJson(rq))
	return rq, nil
}

func ProcessResetQuery(rtrPduModel RtrPduModel) (resetResponses []RtrPduModel, err error) {
	rtrFulls, rtrAsaFulls, sessionId, serialNumber, err := getRtrFullAndSessionIdAndSerialNumberDb()
	if err != nil {
		belogs.Error("ProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	belogs.Debug("ProcessResetQuery(): len(rtrFulls):", len(rtrFulls), " sessionId:", sessionId,
		" serialNumber: ", serialNumber)
	rtrPduModels, err := assembleResetResponses(rtrFulls, rtrAsaFulls, rtrPduModel.GetProtocolVersion(), sessionId, serialNumber)
	if err != nil {
		belogs.Error("ProcessResetQuery(): assembleResetResponses fail: ", err)
		return resetResponses, err
	}
	return rtrPduModels, nil
}

// when len(rtrFull)==0, it is an error with no_data_available
func assembleResetResponses(rtrFulls []model.LabRpkiRtrFull, rtrAsaFulls []model.LabRpkiRtrAsaFull,
	protocolVersion uint8, sessionId uint16, serialNumber uint32) (rtrPduModels []RtrPduModel, err error) {
	belogs.Info("assembleResetResponses(): len(rtrFulls):", len(rtrFulls), " len(rtrAsaFulls):", len(rtrAsaFulls),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	rtrPduModels = make([]RtrPduModel, 0)
	//rtr full from roa rtr
	if protocolVersion == PDU_PROTOCOL_VERSION_0 || protocolVersion == PDU_PROTOCOL_VERSION_1 {
		if len(rtrFulls) > 0 {
			belogs.Debug("assembleResetResponses(): protocolVersion=0 or 1, len(rtrFulls)>0, len(rtrFulls): ", len(rtrFulls),
				"  protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// start response
			cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
			rtrPduModels = append(rtrPduModels, cacheResponseModel)
			belogs.Debug("assembleResetResponses(): protocolVersion=0 or 1, cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

			// rtr full to response
			rtrFullPduModels, err := convertRtrFullsToRtrPduModels(rtrFulls, protocolVersion)
			if err != nil {
				belogs.Error("assembleResetResponses(): protocolVersion=0 or 1, convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrFullPduModels...)
			belogs.Debug("assembleResetResponses(): protocolVersion=0 or 1, len(rtrFullPduModels) : ", len(rtrFullPduModels))

			// end response
			endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
			rtrPduModels = append(rtrPduModels, endOfDataModel)
			belogs.Debug("assembleResetResponses(): protocolVersion=0 or 1, endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

			belogs.Info("assembleResetResponses(): protocolVersion=0 or 1, will send will send Cache Response of all rtr,",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber,
				",  len(rtrFulls): ", len(rtrFulls), ",  len(rtrPduModels):", len(rtrPduModels))
			belogs.Debug("assembleResetResponses(): protocolVersion=0 or 1, rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil

		} else {
			errorReportModel := NewRtrErrorReportModel(protocolVersion, PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)
			rtrPduModels = append(rtrPduModels, errorReportModel)
			belogs.Info("assembleResetResponses(): protocolVersion=0 or 1,there is no rtr this time,  will send errorReport with not_data_available, ",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil
		}
	} else if protocolVersion == PDU_PROTOCOL_VERSION_2 {
		//rtr full from asa rtr
		if len(rtrFulls) > 0 || len(rtrAsaFulls) > 0 {
			belogs.Debug("assembleResetResponses(): protocolVersion=2, len(rtrFulls):", len(rtrFulls), " len(rtrAsaFulls): ", len(rtrAsaFulls),
				"  protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// start response
			cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
			rtrPduModels = append(rtrPduModels, cacheResponseModel)
			belogs.Debug("assembleResetResponses(): cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

			// from rtr full
			rtrFullPduModels, err := convertRtrFullsToRtrPduModels(rtrFulls, protocolVersion)
			if err != nil {
				belogs.Error("assembleResetResponses(): convertRtrFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrFullPduModels...)
			belogs.Debug("assembleResetResponses(): len(rtrFullPduModels) : ", len(rtrFullPduModels))

			// rtr asa full to response
			rtrAsaFullPduModels, err := convertRtrAsaFullsToRtrPduModels(rtrAsaFulls, protocolVersion)
			if err != nil {
				belogs.Error("assembleResetResponses(): convertRtrAsaFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsaFullPduModels...)
			belogs.Debug("assembleResetResponses(): len(rtrAsaFullPduModels) : ", len(rtrAsaFullPduModels))

			// end response
			endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
			rtrPduModels = append(rtrPduModels, endOfDataModel)
			belogs.Debug("assembleResetResponses(): will send all rtrPduModels : ", jsonutil.MarshalJson(rtrPduModels))

			belogs.Info("assembleResetResponses(): protocolVersion=2, will send will send Cache Response of all rtr,",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber,
				",  len(rtrFulls): ", len(rtrFulls), ", len(rtrAsaFulls): ", len(rtrAsaFulls),
				",  len(rtrPduModels):", len(rtrPduModels))
			belogs.Debug("assembleResetResponses(): protocolVersion=2,	rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))

			return rtrPduModels, nil
		} else {
			belogs.Debug("assembleResetResponses(): protocolVersion=2, len(rtrAsaFulls)==0 : just send endofdata,",
				" protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
			rtrPduModels = assembleEndOfDataResponses(protocolVersion, sessionId, serialNumber)

			belogs.Info("assembleResetResponses(): protocolVersion=2,there is no rtr this time,  will send errorReport with not_data_available, ",
				",  receive protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))
			return rtrPduModels, nil
		}
	}

	belogs.Error("assembleResetResponses(): not support protocolVersion, fail: ", protocolVersion)
	return nil, errors.New("protocolVersion is not support")

}

func convertRtrFullToRtrPduModel(rtrFull *model.LabRpkiRtrFull, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrFull.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := NewRtrIpv4PrefixModel(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv4, uint32(rtrFull.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := NewRtrIpv6PrefixModel(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv6, uint32(rtrFull.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}

func convertRtrFullsToRtrPduModels(rtrFulls []model.LabRpkiRtrFull,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrFulls {
		rtrPduModel, err := convertRtrFullToRtrPduModel(&rtrFulls[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrFullsToRtrPduModels(): convertRtrFullToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrFullsToRtrPduModels(): len(rtrFulls): ", len(rtrFulls), " len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrAsaFullsToRtrPduModels(rtrAsaFulls []model.LabRpkiRtrAsaFull,
	protocolVersion uint8) (rtrAsaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrAsaFullsToRtrPduModels(): len(rtrAsaFulls): ", len(rtrAsaFulls), "  protocolVersion:", protocolVersion)

	start := time.Now()
	sameCustomerAsnAfi := make(map[string]*RtrAsaModel, 0)
	rtrAsaPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsaFulls {
		rtrPduModel := NewRtrAsaModelFromDb(protocolVersion, PDU_FLAG_ANNOUNCE,
			rtrAsaFulls[i].AddressFamily, uint32(rtrAsaFulls[i].CustomerAsn))
		key := rtrPduModel.GetKey()
		belogs.Debug("convertRtrAsaFullsToRtrPduModels(): will add key:", key)
		if v, ok := sameCustomerAsnAfi[key]; ok {
			v.AddProviderAsn(uint32(rtrAsaFulls[i].ProviderAsn))
			sameCustomerAsnAfi[key] = v
		} else {
			rtrPduModel.AddProviderAsn(uint32(rtrAsaFulls[i].ProviderAsn))
			sameCustomerAsnAfi[key] = rtrPduModel
		}
	}
	for _, v := range sameCustomerAsnAfi {
		rtrAsaPduModels = append(rtrAsaPduModels, v)
		belogs.Debug("convertRtrAsaFullsToRtrPduModels(): v: ", jsonutil.MarshalJson(v))
	}
	belogs.Info("convertRtrAsaFullsToRtrPduModels(): len(rtrAsaFulls): ", len(rtrAsaFulls),
		" len(rtrAsaPduModels):", len(rtrAsaPduModels), "  time(s):", time.Since(start))
	return rtrAsaPduModels, nil
}
