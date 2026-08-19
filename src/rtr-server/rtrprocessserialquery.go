package rtrserver

import (
	"errors"
	"time"

	rtrcore "github.com/bgpsecurity/rpstir2/rtr-core"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func processSerialQuery(rtrPduModel rtrcore.RtrPduModel) (serialResponses []rtrcore.RtrPduModel, err error) {
	start := time.Now()
	rtrSerialQueryModel, p := rtrPduModel.(*rtrcore.RtrSerialQueryModel)
	if !p {
		belogs.Error("processSerialQuery(): rtrPduModel convert to rtrResetQueryModel fail ")
		return nil, errors.New("processSerialQuery(): rtrPduModel convert to rtrResetQueryModel fail  ")
	}
	clientSessionId := rtrSerialQueryModel.SessionId
	clientSerialNumber := rtrSerialQueryModel.SerialNumber
	belogs.Info("processSerialQuery(): clientSessionId:", clientSessionId, "  clientSerialNum:", clientSerialNumber)

	//
	serialNumbers, err := needResetQuery(clientSessionId, clientSerialNumber)
	belogs.Debug("processSerialQuery(): needReset,   clientSessionId, clientSerialNumber,serialNumbers : ", clientSessionId, clientSerialNumber, serialNumbers)
	if err != nil {
		belogs.Error("processSerialQuery(): needResetQuery fail ,  clientSessionId, clientSerialNumber, err:", clientSessionId, clientSerialNumber, err)
		return nil, errors.New("processSerialQuery(): needResetQuery fail  ")
	}
	belogs.Info("processSerialQuery():  clientSessionId:", clientSessionId, ",  clientSerialNumber:", clientSerialNumber,
		"  server get serialNumbers between client and server: ", jsonutil.MarshalJson(serialNumbers),
		"  time(s):", time.Since(start))

	//
	if len(serialNumbers) == 0 {
		// no new data, so just send End Of Data PDU
		rtrPduModels := rtrcore.AssembleEndOfDataResponses(rtrSerialQueryModel.GetProtocolVersion(), clientSessionId, clientSerialNumber)
		belogs.Info("processSerialQuery(): server get len(serialNumbers) == 0, will just send End Of Data PDU Response,",
			"  clientSessionId: ", clientSessionId, ",  clientSerialNumber:", clientSerialNumber,
			",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
		return rtrPduModels, nil

	} else if len(serialNumbers) > 2 {
		// shloud send Cache Reset PDU Response
		belogs.Debug("processSerialQuery(): server get len(serialNumbers) >2, will send Cache Reset PDU Response,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber, ", len(serialNumbers):", len(serialNumbers))
		rtrPduModels, err := rtrcore.AssembleCacheResetResponses(rtrSerialQueryModel.GetProtocolVersion())
		if err != nil {
			belogs.Error("processSerialQuery(): len(serialNumbers) >2, AssembleCacheResetResponses , fail: ", err)
			return nil, err
		}
		belogs.Info("processSerialQuery(): server get len(serialNumbers) >2, will send Cache Reset PDU Response,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers):", len(serialNumbers), ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
		return rtrPduModels, nil
	} else if len(serialNumbers) > 0 && len(serialNumbers) <= 2 {
		// send Cache Response
		belogs.Debug("processSerialQuery():server get  len(serialNumbers) >0 && <=2 , will send Cache Response of rtr incremental,",
			" clientSessionId: ", clientSessionId, ", clientSerialNumber:", clientSerialNumber,
			", len(serialNumbers): ", len(serialNumbers))
		rtrIncrementals, rtrAsaIncrementals, rtrMoaIncrementals,
			sessionId, serialNumber, err := getRtrXIncrementalDb(clientSerialNumber)
		if err != nil {
			belogs.Error("processSerialQuery(): len(serialNumbers) >0 && <=2,  getRtrXIncrementalDb fail: ", clientSerialNumber, err)
			return nil, err
		}
		belogs.Debug("processSerialQuery(): len(rtrIncrementals):", len(rtrIncrementals),
			"  len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "  len(rtrMoaIncrementals):", len(rtrMoaIncrementals),
			"  sessionId:", sessionId, "  serialNumber:", serialNumber)

		rtrPduModels, err := rtrcore.AssembleSerialResponses(rtrIncrementals, rtrAsaIncrementals, rtrMoaIncrementals,
			rtrSerialQueryModel.GetProtocolVersion(), sessionId, serialNumber)
		if err != nil {
			belogs.Error("processSerialQuery():server get len(serialNumbers) >0 && <=2 , AssembleSerialResponses fail: ", err)
			return nil, err
		}
		belogs.Info("processSerialQuery():server get  len(serialNumbers) >0 && <=2 , will send Cache Response of rtr incremental,",
			"   clientSessionId: ", clientSessionId, "  clientSerialNumber:", clientSerialNumber,
			"   len(serialNumbers): ", len(serialNumbers),
			"   len(rtrIncrementals):", len(rtrIncrementals),
			"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
			"   len(rtrPduModels):", len(rtrPduModels), "  time(s):", time.Since(start))

		return rtrPduModels, nil
	}
	return nil, errors.New("processSerialQuery(): server get serial number from client is err")
}

// 1: check error;  ;
func needResetQuery(clientSessionId uint16, clientSerialNumber uint32) (serialNumbers []uint32, err error) {

	sessionId, err := getSessionIdDb()
	belogs.Debug("needResetQuery(): sessionId:", sessionId, "  clientSessionId:", clientSessionId)
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
