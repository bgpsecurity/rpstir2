package rtrserver

import (
	rtrcore "github.com/bgpsecurity/rpstir2/rtr-core"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func processSerialNotify(protocolVersion uint8) (rtrPduModel rtrcore.RtrPduModel, err error) {

	sessionId, serialNumber, err := getSessionIdAndSerialNumberDb()
	if err != nil {
		belogs.Error("processSerialNotify():getSessionIdAndSerialNumberDb fail:", err)
		rtrError := rtrcore.NewRtrError(
			err,
			false, protocolVersion, rtrcore.PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
			nil, "")
		return rtrPduModel, rtrError
	}

	belogs.Debug("processSerialNotify(): protocolVersion:", protocolVersion, " sessionId:", sessionId, "  serialNumber:", serialNumber)
	rtrSerialNotifyModel := rtrcore.NewRtrSerialNotifyModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("processSerialNotify(): rtrSerialNotifyModel: ", jsonutil.MarshalJson(rtrSerialNotifyModel))
	return rtrSerialNotifyModel, nil

}
