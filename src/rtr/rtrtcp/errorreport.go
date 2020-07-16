package rtrtcp

import (
	"bytes"
	"encoding/binary"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

func ParseToErrorReport(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	/*
		ProtocolVersion        uint8  `json:"protocolVersion"`
		PduType                uint8  `json:"pduType"`
		ErrorCode              uint16 `json:"errorCode"`
		Length                 uint32 `json:"length"`
		LengthOfEncapsulated   uint32 `json:"lengthOfEncapsulated"`
		ErroneousPdu           []byte `json:"erroneousPdu"`
		LengthOfErrorText      uint32 `json:"lengthOfErrorText"`
		ErrorDiagnosticMessage []byte `json:"errorDiagnosticMessage"`
	*/

	var errorCode uint16
	var length uint32
	var lengthOfEncapsulated uint32
	// var erroneousPdu []byte
	var lengthOfErrorText uint32
	//var errorDiagnosticMessage []byte

	// get errorCode
	err = binary.Read(buf, binary.BigEndian, &errorCode)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get errorCode fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get errorCode")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get lengthOfEncapsulated
	err = binary.Read(buf, binary.BigEndian, &lengthOfEncapsulated)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get LengthOfEncapsulated fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get lengthOfEncapsulated")
		return rtrPduModel, rtrError
	}

	// get erroneousPdu
	erroneousPdu := make([]byte, lengthOfEncapsulated)
	err = binary.Read(buf, binary.BigEndian, &erroneousPdu)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get erroneousPdu fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get erroneousPdu")
		return rtrPduModel, rtrError
	}

	// get lengthOfErrorText
	err = binary.Read(buf, binary.BigEndian, &lengthOfErrorText)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get lengthOfErrorText fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get lengthOfErrorText")
		return rtrPduModel, rtrError
	}

	// get errorDiagnosticMessage
	errorDiagnosticMessage := make([]byte, lengthOfErrorText)
	err = binary.Read(buf, binary.BigEndian, &errorDiagnosticMessage)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get erroneousPdu fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get errorDiagnosticMessage")
		return rtrPduModel, rtrError
	}

	sq := rtrmodel.NewRtrErrorReportModel(protocolVersion, errorCode,
		erroneousPdu, errorDiagnosticMessage)
	belogs.Debug("ParseToErrorReport():get PDU_TYPE_ERROR_REPORT ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}

func assembleErrorReportResponse(buf *bytes.Reader, protocolVersion uint8, errorCode uint16,
	errorDiagnosticMessage string) (rtrPduModel rtrmodel.RtrPduModel) {

	buf.Seek(0, 0)
	erroneousPdu := make([]byte, buf.Size())
	buf.Read(erroneousPdu)

	errorReportModel := rtrmodel.NewRtrErrorReportModel(protocolVersion, errorCode,
		erroneousPdu, []byte(errorDiagnosticMessage))
	belogs.Debug("AssembleErrorReportResponses(): errorReportModel : ", jsonutil.MarshalJson(errorReportModel))

	return errorReportModel
}
