package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"

	db "rtr/db"
	rtrmodel "rtr/model"
)

func ParseToSerialNotify(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialNotify(): PDU_TYPE_SERIAL_NOTIFY get sessionId fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get SessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialNotify(): PDU_TYPE_SERIAL_NOTIFY get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 12 {
		belogs.Error("ParseToSerialNotify():PDU_TYPE_SERIAL_NOTIFY,  length must be 12 ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is SERIAL NOTIFY,  length must be 12"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialNotify(): PDU_TYPE_SERIAL_NOTIFY get serialNumber fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}

	rtrPduModel = rtrmodel.NewRtrSerialNotifyModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialNotify():get PDU_TYPE_SERIAL_NOTIFY ", buf, jsonutil.MarshalJson(rtrPduModel))
	return rtrPduModel, nil
}

func ProcessSerialNotify(protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	sessionId, serialNumber, err := db.GetSessionIdAndSerialNumber()
	if err != nil {
		belogs.Error("ProcessSerialNotify():GetSessionIdAndSerialNumber fail:", err)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
			nil, "")
		return rtrPduModel, rtrError
	}

	rtrSerialNotifyModel := rtrmodel.NewRtrSerialNotifyModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ProcessSerialNotify(): rtrSerialNotifyModel : ", jsonutil.MarshalJson(rtrSerialNotifyModel))
	return rtrSerialNotifyModel, nil

}
