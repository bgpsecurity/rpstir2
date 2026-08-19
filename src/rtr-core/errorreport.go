package rtrcore

import (
	"bytes"
	"encoding/binary"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrError struct {
	Err error `json:"err"`
	// if get error pdu ,do not send response
	NeedSendResponse bool `json:"needSendResponse"`

	ProtocolVersion        uint8  `json:"protocolVersion"`
	ErrorCode              uint16 `json:"errorCode"`
	ErroneousPdu           []byte `json:"erroneousPdu"`
	ErrorDiagnosticMessage []byte `json:"errorDiagnosticMessage"`
}

func NewRtrError(err error, needSendResponse bool, protocolVersion uint8, errorCode uint16,
	buf *bytes.Reader, errorDiagnosticMessage string) *RtrError {
	var erroneousPdu []byte
	if buf != nil {
		buf.Seek(0, 0)
		erroneousPdu = make([]byte, buf.Size())
		buf.Read(erroneousPdu)
	} else {
		erroneousPdu = nil
	}

	rtrError := &RtrError{
		Err:                    err,
		NeedSendResponse:       needSendResponse,
		ProtocolVersion:        protocolVersion,
		ErrorCode:              errorCode,
		ErroneousPdu:           erroneousPdu,
		ErrorDiagnosticMessage: []byte(errorDiagnosticMessage),
	}

	return rtrError
}

func (p *RtrError) Error() string {
	return p.Err.Error()
}
func (p *RtrError) Unwrap() error {
	return p.Err
}

type RtrErrorReportModel struct {
	ProtocolVersion        uint8  `json:"protocolVersion"`
	PduType                uint8  `json:"pduType"`
	ErrorCode              uint16 `json:"errorCode"`
	Length                 uint32 `json:"length"`
	LengthOfEncapsulated   uint32 `json:"lengthOfEncapsulated"`
	ErroneousPdu           []byte `json:"erroneousPdu"`
	LengthOfErrorText      uint32 `json:"lengthOfErrorText"`
	ErrorDiagnosticMessage []byte `json:"errorDiagnosticMessage"`
}

// erroneousPdu and errorDiagnosticMessage can be nil
func NewRtrErrorReportModel(protocolVersion uint8, errorCode uint16,
	erroneousPdu []byte, errorDiagnosticMessage []byte) *RtrErrorReportModel {
	erm := &RtrErrorReportModel{PduType: PDU_TYPE_ERROR_REPORT}
	erm.ProtocolVersion = protocolVersion
	erm.ErrorCode = errorCode
	erm.LengthOfEncapsulated = uint32(len(erroneousPdu))
	erm.ErroneousPdu = erroneousPdu
	erm.LengthOfErrorText = uint32(len(errorDiagnosticMessage))
	erm.ErrorDiagnosticMessage = errorDiagnosticMessage
	// (protocolversion+pdutype+errorCode)+length + lengthofencapsulatedpdu + ErroneousPDU + LengthOfErrorText + errorDiagnosticMessage
	erm.Length = 4 + 4 + 4 + uint32(len(erroneousPdu)) + 4 + uint32(len(errorDiagnosticMessage))

	return erm
}

// erroneousPdu and errorDiagnosticMessage can be nil
func NewRtrErrorReportModelByRtrError(rtrError *RtrError) *RtrErrorReportModel {

	return NewRtrErrorReportModel(rtrError.ProtocolVersion, rtrError.ErrorCode,
		rtrError.ErroneousPdu, rtrError.ErrorDiagnosticMessage)
}

func (p *RtrErrorReportModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.ErrorCode)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.LengthOfEncapsulated)
	if len(p.ErroneousPdu) > 0 {
		binary.Write(wr, binary.BigEndian, p.ErroneousPdu)
	}
	binary.Write(wr, binary.BigEndian, p.LengthOfErrorText)
	if len(p.ErrorDiagnosticMessage) > 0 {
		binary.Write(wr, binary.BigEndian, p.ErrorDiagnosticMessage)
	}
	return wr.Bytes()
}

func (p *RtrErrorReportModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrErrorReportModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrErrorReportModel) GetPduType() uint8 {
	return p.PduType
}
func ParseToErrorReport(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
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
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get errorCode fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			false, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get errorCode")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			false, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get lengthOfEncapsulated
	err = binary.Read(buf, binary.BigEndian, &lengthOfEncapsulated)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get LengthOfEncapsulated fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			false, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get lengthOfEncapsulated")
		return rtrPduModel, rtrError
	}

	// get erroneousPdu
	erroneousPdu := make([]byte, lengthOfEncapsulated)
	err = binary.Read(buf, binary.BigEndian, &erroneousPdu)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get erroneousPdu fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			false, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get erroneousPdu")
		return rtrPduModel, rtrError
	}

	// get lengthOfErrorText
	err = binary.Read(buf, binary.BigEndian, &lengthOfErrorText)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get lengthOfErrorText fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			false, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get lengthOfErrorText")
		return rtrPduModel, rtrError
	}

	// get errorDiagnosticMessage
	errorDiagnosticMessage := make([]byte, lengthOfErrorText)
	err = binary.Read(buf, binary.BigEndian, &errorDiagnosticMessage)
	if err != nil {
		belogs.Error("ParseToErrorReport(): PDU_TYPE_ERROR_REPORT get errorDiagnosticMessage fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			false, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get errorDiagnosticMessage")
		return rtrPduModel, rtrError
	}

	sq := NewRtrErrorReportModel(protocolVersion, errorCode,
		erroneousPdu, errorDiagnosticMessage)
	belogs.Error("ParseToErrorReport(): RtrErrorReportModel: erroneousPdu:", erroneousPdu,
		"  errorDiagnosticMessage:", string(errorDiagnosticMessage))
	belogs.Debug("ParseToErrorReport():get PDU_TYPE_ERROR_REPORT, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}

func AssembleErrorReportResponse(buf *bytes.Reader, protocolVersion uint8, errorCode uint16,
	errorDiagnosticMessage string) (rtrPduModel RtrPduModel) {

	buf.Seek(0, 0)
	erroneousPdu := make([]byte, buf.Size())
	buf.Read(erroneousPdu)

	errorReportModel := NewRtrErrorReportModel(protocolVersion, errorCode,
		erroneousPdu, []byte(errorDiagnosticMessage))
	belogs.Debug("AssembleErrorReportResponses(): errorReportModel : ", jsonutil.MarshalJson(errorReportModel))

	return errorReportModel
}
