package rtrserver

import (
	"bytes"
	"errors"
	"net"
	"time"

	rtrcore "github.com/bgpsecurity/rpstir2/rtr-core"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

func ProcessRtrPduModel(buf *bytes.Reader, rtrPduModel rtrcore.RtrPduModel) (rtrResponse []rtrcore.RtrPduModel, err error) {

	pduType := rtrPduModel.GetPduType()
	belogs.Debug("processRtrPduModel():pduType: ", pduType)
	switch pduType {
	case rtrcore.PDU_TYPE_SERIAL_QUERY:
		serialResponse, err := processSerialQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): processSerialQuery fail:", err)
			rtrError := rtrcore.NewRtrError(
				err,
				false, rtrPduModel.GetProtocolVersion(), rtrcore.PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
				buf, "Fail to get pdu type")
			return nil, rtrError
		}
		belogs.Debug("processRtrPduModel():len(serialResponse): ", len(serialResponse), jsonutil.MarshalJson(serialResponse))
		return serialResponse, nil
	case rtrcore.PDU_TYPE_RESET_QUERY:
		resetResponse, err := processResetQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): processResetQuery fail: ", err)
			rtrError := rtrcore.NewRtrError(
				err,
				false, rtrPduModel.GetProtocolVersion(), rtrcore.PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
				buf, "Fail to get pdu type")
			return nil, rtrError
		}
		belogs.Debug("processRtrPduModel():len(resetResponse): ", len(resetResponse))
		return resetResponse, nil
	case rtrcore.PDU_TYPE_ERROR_REPORT:
		belogs.Info("processRtrPduModel():no need to process error report: ", jsonutil.MarshalJson(rtrPduModel))
		rtrResponse = make([]rtrcore.RtrPduModel, 0)
		return rtrResponse, nil
	default:
		belogs.Error("processRtrPduModel():pdutype should not recevie by rtr server, is ", pduType)
		rtrError := rtrcore.NewRtrError(
			err,
			false, rtrPduModel.GetProtocolVersion(), rtrcore.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pdu type")
		return nil, rtrError

	}
}

func SendResponses(conn *net.TCPConn, rtrPduModelResponses []rtrcore.RtrPduModel) (err error) {
	start := time.Now()
	// batchId Ms is unique id in the same batch, get from start time
	batchId := start.UnixNano() / 1e6
	sendIntervalMs := conf.Int("rtr::sendIntervalMs")
	for _, one := range rtrPduModelResponses {
		sendBytes := one.Bytes()
		//conn.SetWriteBuffer(len(sendBytes))
		n, err := conn.Write(sendBytes)
		if err != nil {
			belogs.Debug("SendResponses():  conn.Write() fail,  ", jsonutil.MarshalJson(one), n, err)
			return err
		}
		belogs.Debug("SendResponses():send batchId:", batchId, ", rtrPduModel:", jsonutil.MarshalJson(one),
			", len(sendBytes):", len(sendBytes), ",  sendBytes:\n"+convert.PrintBytes(sendBytes, 8))

		// avoid tcp sticky packets
		if sendIntervalMs > 0 {
			time.Sleep(time.Duration(sendIntervalMs) * time.Millisecond)
		}
	}
	belogs.Debug("SendResponses(): send len(packets):", len(rtrPduModelResponses), ",   time(s):", time.Since(start))
	return nil
}
func SendErrorResponse(conn *net.TCPConn, err error) (er error) {
	belogs.Debug("SendErrorResponse():  err: ", err)
	var rtrError *rtrcore.RtrError
	if errors.As(err, &rtrError) && rtrError.NeedSendResponse {
		belogs.Debug("SendErrorResponse():will send rtr Error: ", jsonutil.MarshalJson(rtrError))
		return sendErrorResponse(conn, rtrError)
	}
	return nil
}

func sendErrorResponse(conn *net.TCPConn, rtrError *rtrcore.RtrError) (err error) {
	start := time.Now()
	rtrErrorReportModel := rtrcore.NewRtrErrorReportModelByRtrError(rtrError)
	sendBytes := rtrErrorReportModel.Bytes()
	belogs.Debug("SendResponses(): send by conn:", convert.Bytes2String(sendBytes))
	//conn.SetWriteBuffer(len(sendBytes))
	n, err := conn.Write(sendBytes)
	if err != nil {
		belogs.Debug("SendResponses():  conn.Write() fail,  ", jsonutil.MarshalJson(rtrErrorReportModel), err)
		return err
	}
	belogs.Info("SendResponses(): send n, packets:", n, jsonutil.MarshalJson(rtrErrorReportModel), ",   time(s):", time.Since(start))
	return nil
}
