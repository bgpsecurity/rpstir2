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

func ParseToResetQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	var zero16 uint16
	var length uint32

	// get zero16
	err = binary.Read(buf, binary.BigEndian, &zero16)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get zero fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToResetQuery():PDU_TYPE_RESET_QUERY,  length must be 8 ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is SERIAL QUERY, length must be 8"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	rq := rtrmodel.NewRtrResetQueryModel(protocolVersion)
	belogs.Debug("ParseToResetQuery():get PDU_TYPE_RESET_QUERY ", buf, jsonutil.MarshalJson(rq))
	return rq, nil
}

func ProcessResetQuery(rtrPduModel rtrmodel.RtrPduModel) (resetResponses []rtrmodel.RtrPduModel, err error) {
	rtrFulls, sessionId, serialNumber, err := db.GetRtrFullAndSessionIdAndSerialNumber()
	if err != nil {
		belogs.Error("ProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	belogs.Debug("ProcessResetQuery(): rtrFulls, sessionId, serialNumber: ", len(rtrFulls), sessionId, serialNumber)
	rtrPduModels, err := assembleResetResponses(rtrFulls, rtrPduModel.GetProtocolVersion(), sessionId, serialNumber)
	if err != nil {
		belogs.Error("ProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	return rtrPduModels, nil
}

// when len(rtrFull)==0, it is an error with no_data_available
func assembleResetResponses(rtrFulls []model.LabRpkiRtrFull, protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModels []rtrmodel.RtrPduModel, err error) {
	rtrPduModels = make([]rtrmodel.RtrPduModel, 0)

	if len(rtrFulls) > 0 {
		belogs.Debug("assembleResetResponses(): will send will send Cache Response of all rtr,",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ", len(rtr): ", len(rtrFulls))

		cacheResponseModel := rtrmodel.NewRtrCacheResponseModel(protocolVersion, sessionId)
		belogs.Debug("assembleResetResponses(): cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

		rtrPduModels = append(rtrPduModels, cacheResponseModel)
		for i, _ := range rtrFulls {
			rtrPduModel, err := convertRtrFullToRtrPduModel(&rtrFulls[i], protocolVersion)
			if err != nil {
				belogs.Error("assembleResetResponses(): convertRtrFullToRtrPduModel fail: ", err)
				return rtrPduModels, err
			}
			belogs.Debug("assembleResetResponses(): rtrPduModel : ", jsonutil.MarshalJson(rtrPduModel))

			rtrPduModels = append(rtrPduModels, rtrPduModel)
		}

		endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		belogs.Debug("assembleResetResponses(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

		rtrPduModels = append(rtrPduModels, endOfDataModel)
		belogs.Info("assembleResetResponses(): will send will send Cache Response of all rtr,",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber,
			", len(rtr): ", len(rtrFulls), ",  len(rtrPduModels):", len(rtrPduModels))

	} else {
		belogs.Debug("assembleResetResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber)
		errorReportModel := rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)

		rtrPduModels = append(rtrPduModels, errorReportModel)
		belogs.Info("assembleResetResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))

	}
	return rtrPduModels, nil

}

func convertRtrFullToRtrPduModel(rtrFull *model.LabRpkiRtrFull, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrFull.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := rtrmodel.NewRtrIpv4PrefixModel(protocolVersion, 1, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv4, uint32(rtrFull.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := rtrmodel.NewRtrIpv6PrefixModel(protocolVersion, 1, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv6, uint32(rtrFull.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}
