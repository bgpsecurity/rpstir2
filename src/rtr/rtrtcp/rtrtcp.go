package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

// rtrPduModel:
func ParseToRtrPduModel(buf *bytes.Reader) (rtrPduModel rtrmodel.RtrPduModel, err error) {

	// get length
	if buf.Size() < rtrmodel.PDU_TYPE_MIN_LEN {
		belogs.Error("ParseToRtrPduModel(): recv byte's length is too small: ", buf.Size())
		rtrError := rtrmodel.NewRtrError(
			errors.New("length of receive bytes is too small"),
			false, rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_INVALID_REQUEST,
			buf, "")
		return rtrPduModel, rtrError
	}

	// get protocolVersion, pduType
	protocolVersion, pduType, err := parseProtocolVersionAndPduType(buf)
	if err != nil {
		belogs.Error("ParseToRtrPduModel():parseProtocolVersionAndPduType err: ", err)
		return rtrPduModel, err
	}

	belogs.Debug("ParseToRtrPduModel():  protocolVersion, pduType:", protocolVersion, pduType)
	switch pduType {
	case rtrmodel.PDU_TYPE_SERIAL_NOTIFY:
		return ParseToSerialNotify(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_SERIAL_QUERY:
		return ParseToSerialQuery(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_RESET_QUERY:
		return ParseToResetQuery(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_CACHE_RESPONSE:
		return ParseToCacheResponse(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_IPV4_PREFIX:
		return ParseToIpv4Prefix(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_IPV6_PREFIX:
		return ParseToIpv6Prefix(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_END_OF_DATA:
		return ParseToEndOfData(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_CACHE_RESET:
		return ParseToCacheReset(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_ROUTER_KEY:
		return ParseToRouterKey(buf, protocolVersion)

	case rtrmodel.PDU_TYPE_ERROR_REPORT:
		return ParseToErrorReport(buf, protocolVersion)

	default:
		belogs.Error("parseToRtrPduModel():received bytes cannot be parse to rtr's pdu,  pduType:", pduType)
		rtrError := rtrmodel.NewRtrError(
			errors.New("received bytes cannot be parse to rtr's pdu, is "+strconv.Itoa(int(pduType))),
			false, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pdu type")
		return rtrPduModel, rtrError
	}

}

func ProcessRtrPduModel(buf *bytes.Reader, rtrPduModel rtrmodel.RtrPduModel) (rtrResponse []rtrmodel.RtrPduModel, err error) {

	pduType := rtrPduModel.GetPduType()
	belogs.Debug("processRtrPduModel():pduType: ", pduType)
	switch pduType {
	case rtrmodel.PDU_TYPE_SERIAL_QUERY:
		serialResponse, err := ProcessSerialQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): ProcessSerialQuery fail: ", err)
			rtrError := rtrmodel.NewRtrError(
				err,
				false, rtrPduModel.GetProtocolVersion(), rtrmodel.PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
				buf, "Fail to get pdu type")
			return nil, rtrError
		}
		belogs.Debug("processRtrPduModel():serialResponse: ", jsonutil.MarshalJson(serialResponse))
		return serialResponse, nil
	case rtrmodel.PDU_TYPE_RESET_QUERY:
		resetResponse, err := ProcessResetQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): ProcessResetQuery fail: ", err)
			rtrError := rtrmodel.NewRtrError(
				err,
				false, rtrPduModel.GetProtocolVersion(), rtrmodel.PDU_TYPE_ERROR_CODE_INTERNAL_ERROR,
				buf, "Fail to get pdu type")
			return nil, rtrError
		}
		belogs.Debug("processRtrPduModel():resetResponse: ", jsonutil.MarshalJson(resetResponse))
		return resetResponse, nil
	default:
		belogs.Error("processRtrPduModel():pdutype should not recevie by rtr server, is ", pduType)
		rtrError := rtrmodel.NewRtrError(
			err,
			false, rtrPduModel.GetProtocolVersion(), rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pdu type")
		return nil, rtrError

	}
}

func SendResponses(conn *net.TCPConn, rtrPduModelResponses []rtrmodel.RtrPduModel) (err error) {
	start := time.Now()
	// batchId Ms is unique id in the same batch, get from start time
	batchId := start.UnixNano() / 1e6
	sendIntervalMs := conf.Int("rtr:sendIntervalMs")
	for _, one := range rtrPduModelResponses {
		sendBytes := one.Bytes()
		belogs.Debug("sendResponses(): send by conn :\r\n", convert.Bytes2String(sendBytes))
		//conn.SetWriteBuffer(len(sendBytes))
		n, err := conn.Write(sendBytes)
		if err != nil {
			belogs.Debug("sendResponses():  conn.Write() fail,  ", jsonutil.MarshalJson(one), err)
			return err
		}
		belogs.Info("SendResponses():send batchId:", batchId, ", rtrPduModel:", jsonutil.MarshalJson(one),
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
	var rtrError *rtrmodel.RtrError
	if errors.As(err, &rtrError) && rtrError.NeedSendResponse {
		belogs.Debug("SendErrorResponse():will send rtr Error: ", jsonutil.MarshalJson(rtrError))
		return sendErrorResponse(conn, rtrError)
	}
	return nil
}

func sendErrorResponse(conn *net.TCPConn, rtrError *rtrmodel.RtrError) (err error) {
	start := time.Now()
	rtrErrorReportModel := rtrmodel.NewRtrErrorReportModelByRtrError(rtrError)
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
		rtrError := rtrmodel.NewRtrError(
			err,
			true, rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get protocolVersion")
		return 0, 0, rtrError
	}
	if protocolVersion != 0 && protocolVersion != 1 {
		belogs.Error("parseToPduModel(): protocolVersion is illegal: ", buf, protocolVersion)
		rtrError := rtrmodel.NewRtrError(
			errors.New("protocolVersion is neigher 0 nor 1, "+strconv.Itoa(int(protocolVersion))),
			true, rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get protocolVersion")
		return 0, 0, rtrError
	}

	// get pdu type
	err = binary.Read(buf, binary.BigEndian, &pduType)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	if pduType > rtrmodel.PDU_TYPE_ERROR_REPORT {
		belogs.Error("parseToPduModel(): pduType is illegal: ", buf, pduType)
		rtrError := rtrmodel.NewRtrError(
			errors.New("get Itoa is error "+strconv.Itoa(int(pduType))),
			true, rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	if pduType == rtrmodel.PDU_TYPE_ROUTER_KEY && protocolVersion == 0 {
		belogs.Error("parseToPduModel():pduType is PDU_TYPE_ROUTER_KEY,  protocolVersion must be 1 ", buf, pduType, protocolVersion)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is ROUTER KEY,  protocolVersion must be 1"),
			true, rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION,
			buf, "Fail to get pduType")
		return 0, 0, rtrError
	}
	belogs.Debug("parseToPduModel():protocolVersion is ", protocolVersion, "  pduType is ", pduType)
	return protocolVersion, pduType, nil
}
