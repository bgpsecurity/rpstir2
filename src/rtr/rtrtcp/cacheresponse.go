package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

func ParseToCacheResponse(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	var sessionId uint16
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToCacheResponse(): PDU_TYPE_CACHE_RESPONSE get sessionId fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get sessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToCacheResponse(): PDU_TYPE_CACHE_RESPONSE get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToCacheResponse():PDU_TYPE_CACHE_RESPONSE,  length must be 8 ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is CACHE RESPONSE, length must be 8"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	sq := rtrmodel.NewRtrCacheResponseModel(protocolVersion, sessionId)
	belogs.Debug("ParseToCacheResponse():get PDU_TYPE_CACHE_RESPONSE ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
