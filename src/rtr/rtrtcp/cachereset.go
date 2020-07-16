package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

func ParseToCacheReset(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	var zero uint16
	var length uint32

	// get zero
	err = binary.Read(buf, binary.BigEndian, &zero)
	if err != nil {
		belogs.Error("ParseToCacheReset(): PDU_TYPE_CACHE_RESET get zero fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToCacheReset(): PDU_TYPE_CACHE_RESET get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToCacheReset():PDU_TYPE_CACHE_RESET,  length must be 8 ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is CACHE RESET, length must be 8"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}
	sq := rtrmodel.NewRtrCacheResetModel(protocolVersion)
	belogs.Debug("ParseToCacheReset():get PDU_TYPE_CACHE_RESET ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
func assembleCacheResetResponses(protocolVersion uint8) (rtrPduModels []rtrmodel.RtrPduModel, err error) {
	rtrPduModels = make([]rtrmodel.RtrPduModel, 0)
	cacheResetResponseModel := rtrmodel.NewRtrCacheResetModel(protocolVersion)
	belogs.Debug("assembleCacheResetResponses(): cacheResetResponseModel : ", jsonutil.MarshalJson(cacheResetResponseModel))
	rtrPduModels = append(rtrPduModels, cacheResetResponseModel)
	return rtrPduModels, nil

}
