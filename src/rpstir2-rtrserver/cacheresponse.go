package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func ParseToCacheResponse(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var sessionId uint16
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToCacheResponse(): PDU_TYPE_CACHE_RESPONSE get sessionId fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get sessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToCacheResponse(): PDU_TYPE_CACHE_RESPONSE get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToCacheResponse():PDU_TYPE_CACHE_RESPONSE,  length must be 8, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is CACHE RESPONSE, length must be 8"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	sq := NewRtrCacheResponseModel(protocolVersion, sessionId)
	belogs.Debug("ParseToCacheResponse():get PDU_TYPE_CACHE_RESPONSE, buf:", buf, " sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
