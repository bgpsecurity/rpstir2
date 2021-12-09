package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

// rtrPduModel:
func ParseToRtrPduModel(buf *bytes.Reader) (rtrPduModel RtrPduModel, err error) {

	// get length
	if buf.Size() < PDU_TYPE_MIN_LEN {
		belogs.Error("ParseToRtrPduModel(): recv byte's length is too small: ", buf.Size())
		rtrError := NewRtrError(
			errors.New("length of receive bytes is too small"),
			false, PROTOCOL_VERSION_0, PDU_TYPE_ERROR_CODE_INVALID_REQUEST,
			buf, "")
		return rtrPduModel, rtrError
	}

	// get protocolVersion, pduType
	protocolVersion, pduType, err := parseProtocolVersionAndPduType(buf)
	if err != nil {
		belogs.Error("ParseToRtrPduModel():parseProtocolVersionAndPduType err: ", err)
		return rtrPduModel, err
	}

	belogs.Info("ParseToRtrPduModel():  protocolVersion, pduType:", protocolVersion, pduType)
	switch pduType {
	case PDU_TYPE_SERIAL_NOTIFY:
		return ParseToSerialNotify(buf, protocolVersion)

	case PDU_TYPE_SERIAL_QUERY:
		return ParseToSerialQuery(buf, protocolVersion)

	case PDU_TYPE_RESET_QUERY:
		return ParseToResetQuery(buf, protocolVersion)

	case PDU_TYPE_CACHE_RESPONSE:
		return ParseToCacheResponse(buf, protocolVersion)

	case PDU_TYPE_IPV4_PREFIX:
		return ParseToIpv4Prefix(buf, protocolVersion)

	case PDU_TYPE_IPV6_PREFIX:
		return ParseToIpv6Prefix(buf, protocolVersion)

	case PDU_TYPE_END_OF_DATA:
		return ParseToEndOfData(buf, protocolVersion)

	case PDU_TYPE_CACHE_RESET:
		return ParseToCacheReset(buf, protocolVersion)

	case PDU_TYPE_ROUTER_KEY:
		return ParseToRouterKey(buf, protocolVersion)

	case PDU_TYPE_ERROR_REPORT:
		return ParseToErrorReport(buf, protocolVersion)

	default:
		belogs.Error("parseToRtrPduModel():received bytes cannot be parse to rtr's pdu,  pduType:", pduType)
		rtrError := NewRtrError(
			errors.New("received bytes cannot be parse to rtr's pdu, is "+strconv.Itoa(int(pduType))),
			false, protocolVersion, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pdu type")
		return rtrPduModel, rtrError
	}

}

func ProcessRtrPduModel(buf *bytes.Reader, rtrPduModel RtrPduModel) (rtrResponse []RtrPduModel, err error) {

	pduType := rtrPduModel.GetPduType()
	belogs.Debug("processRtrPduModel():pduType: ", pduType)
	switch pduType {
	case PDU_TYPE_SERIAL_QUERY:
		serialResponse, err := ProcessSerialQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): ProcessSerialQuery fail: ", err)
			rtrError := NewRtrError(
				err,
				false, rtrPduModel.GetProtocolVersion(), PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
				buf, "Fail to get pdu type")
			return nil, rtrError
		}
		belogs.Debug("processRtrPduModel():len(serialResponse): ", len(serialResponse), jsonutil.MarshalJson(serialResponse))
		return serialResponse, nil
	case PDU_TYPE_RESET_QUERY:
		resetResponse, err := ProcessResetQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): ProcessResetQuery fail: ", err)
			rtrError := NewRtrError(
				err,
				false, rtrPduModel.GetProtocolVersion(), PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
				buf, "Fail to get pdu type")
			return nil, rtrError
		}
		belogs.Debug("processRtrPduModel():len(resetResponse): ", len(resetResponse))
		return resetResponse, nil
	case PDU_TYPE_ERROR_REPORT:
		belogs.Info("processRtrPduModel():no need to process error report: ", jsonutil.MarshalJson(rtrPduModel))
		rtrResponse = make([]RtrPduModel, 0)
		return rtrResponse, nil
	default:
		belogs.Error("processRtrPduModel():pdutype should not recevie by rtr server, is ", pduType)
		rtrError := NewRtrError(
			err,
			false, rtrPduModel.GetProtocolVersion(), PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pdu type")
		return nil, rtrError

	}
}

func SendResponses(conn *net.TCPConn, rtrPduModelResponses []RtrPduModel) (err error) {
	start := time.Now()
	// batchId Ms is unique id in the same batch, get from start time
	batchId := start.UnixNano() / 1e6
	sendIntervalMs := conf.Int("rtr:sendIntervalMs")
	for _, one := range rtrPduModelResponses {
		sendBytes := one.Bytes()
		//belogs.Debug("sendResponses(): send by conn :\r\n", convert.Bytes2String(sendBytes))
		//conn.SetWriteBuffer(len(sendBytes))
		n, err := conn.Write(sendBytes)
		if err != nil {
			belogs.Debug("sendResponses():  conn.Write() fail,  ", jsonutil.MarshalJson(one), n, err)
			return err
		}
		belogs.Debug("SendResponses():send batchId:", batchId, ", rtrPduModel:", jsonutil.MarshalJson(one),
			", should send len(sendBytes):", len(sendBytes), ",  actual send Bytes n:", n)

		// avoid tcp sticky packets
		if sendIntervalMs > 0 {
			time.Sleep(time.Duration(sendIntervalMs) * time.Microsecond)
		}
	}
	belogs.Debug("SendResponses(): send len(packets):", len(rtrPduModelResponses), ",   time(s):", time.Now().Sub(start).Seconds())
	return nil
}
func SendErrorResponse(conn *net.TCPConn, err error) (er error) {
	belogs.Debug("SendErrorResponse():  err: ", err)
	var rtrError *RtrError
	if errors.As(err, &rtrError) && rtrError.NeedSendResponse {
		belogs.Debug("SendErrorResponse():will send rtr Error: ", jsonutil.MarshalJson(rtrError))
		return sendErrorResponse(conn, rtrError)
	}
	return nil
}

func sendErrorResponse(conn *net.TCPConn, rtrError *RtrError) (err error) {
	start := time.Now()
	rtrErrorReportModel := NewRtrErrorReportModelByRtrError(rtrError)
	sendBytes := rtrErrorReportModel.Bytes()
	belogs.Debug("sendResponses(): send by conn :\r\n", convert.Bytes2String(sendBytes))
	//conn.SetWriteBuffer(len(sendBytes))
	n, err := conn.Write(sendBytes)
	if err != nil {
		belogs.Debug("sendResponses():  conn.Write() fail,  ", jsonutil.MarshalJson(rtrErrorReportModel), err)
		return err
	}
	belogs.Info("SendResponses(): send n, packets:", n, jsonutil.MarshalJson(rtrErrorReportModel), ",   time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func parseProtocolVersionAndPduType(buf *bytes.Reader) (protocolVersion, pduType uint8, err error) {

	// get protocol version
	err = binary.Read(buf, binary.BigEndian, &protocolVersion)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, PROTOCOL_VERSION_0, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get protocolVersion")
		return 0, 0, rtrError
	}
	if protocolVersion != 0 && protocolVersion != 1 {
		belogs.Error("parseToPduModel(): protocolVersion is illegal: ", buf, protocolVersion)
		rtrError := NewRtrError(
			errors.New("protocolVersion is neigher 0 nor 1, "+strconv.Itoa(int(protocolVersion))),
			true, PROTOCOL_VERSION_0, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get protocolVersion")
		return 0, 0, rtrError
	}

	// get pdu type
	err = binary.Read(buf, binary.BigEndian, &pduType)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail: ", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	if pduType > PDU_TYPE_ERROR_REPORT {
		belogs.Error("parseToPduModel(): pduType is illegal: ", buf, pduType)
		rtrError := NewRtrError(
			errors.New("get Itoa is error "+strconv.Itoa(int(pduType))),
			true, PROTOCOL_VERSION_0, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	if pduType == PDU_TYPE_ROUTER_KEY && protocolVersion == 0 {
		belogs.Error("parseToPduModel():pduType is PDU_TYPE_ROUTER_KEY,  protocolVersion must be 1 ", buf, pduType, protocolVersion)
		rtrError := NewRtrError(
			errors.New("pduType is ROUTER KEY,  protocolVersion must be 1"),
			true, PROTOCOL_VERSION_0, PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	belogs.Debug("parseToPduModel():protocolVersion is ", protocolVersion, "  pduType is ", pduType)
	return protocolVersion, pduType, nil
}
