package rtrserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

func RtrServerProcess(conn net.Conn) {
	defer conn.Close()

	belogs.Info("RtrServerProcess(): server get new client: ", conn.RemoteAddr())
	buffer := make([]byte, 8000)
	//get connection, and process
	for {
		n, err := conn.Read(buffer)
		belogs.Debug("RtrServerProcess(): read buffer length: ", n)
		if err != nil {
			if err != io.EOF {
				belogs.Error("RtrServerProcess():  read buffer fail: ", err)
			} else {
				belogs.Debug("RtrServerProcess():  client close: ", err)
			}
			return
		}
		if n == 0 {
			continue
		}

		recvByte := buffer[0:n]
		belogs.Debug("RtrServerProcess(): read buffer: ", convert.Bytes2String(recvByte))

		// parse []byte --> rtrpdumodel
		rtrPduModel, err := parseToRtrPduModel(recvByte)
		belogs.Debug("RtrServerProcess():after parseToRtrPduModel, rtrPduModel:", rtrPduModel.GetProtocolVersion(), rtrPduModel.GetPduType())
		if err != nil {
			belogs.Error("RtrServerProcess():  parseToRtrPduModel fail: ", err)
			return
		}

		// process rtrpdumodel --> response rtrpdumodels
		rtrPduModelResponses, err := processRtrPduModel(rtrPduModel)
		belogs.Debug("RtrServerProcess():after processRtrPduModel, rtrPduModelResponses:", jsonutil.MarshalJson(rtrPduModelResponses))
		if err != nil {
			belogs.Error("RtrServerProcess():  processRtrPduModel fail: ", err)
			return
		}

		// send response rtrpdumodels
		err = sendResponses(conn, &rtrPduModelResponses)
		if err != nil {
			belogs.Error("RtrServerProcess():  senRtrPduModelResponses fail: ", err)

			// send internal error

			return
		}
	}
}

func parseProtocolVersionAndPduType(buf *bytes.Reader) (rtrErrorReportPduModel rtrmodel.RtrPduModel,
	protocolVersion, pduType uint8, err error) {

	err = binary.Read(buf, binary.BigEndian, &protocolVersion)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail: ", err)
		return rtrmodel.NewRtrErrorReportModel(rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION, nil, nil),
			protocolVersion, pduType, err
	}
	if protocolVersion != 0 && protocolVersion != 1 {
		belogs.Error("parseToPduModel(): protocolVersion is illegal: ", protocolVersion)
		return rtrmodel.NewRtrErrorReportModel(rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION, nil, nil),
			protocolVersion, pduType, errors.New("get protocolVersion is error " + strconv.Itoa(int(protocolVersion)))
	}
	err = binary.Read(buf, binary.BigEndian, &pduType)
	if err != nil {
		belogs.Error("parseToPduModel(): get protocolVersion from recvByte fail: ", err)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE, nil, nil),
			protocolVersion, pduType, err
	}
	if pduType > rtrmodel.PDU_TYPE_ERROR_REPORT {
		belogs.Error("parseToPduModel(): pduType is illegal: ", pduType)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE, nil, nil),
			protocolVersion, pduType, errors.New("get Itoa is error " + strconv.Itoa(int(pduType)))
	}
	if pduType == rtrmodel.PDU_TYPE_ROUTER_KEY && protocolVersion == 0 {
		belogs.Error("parseToPduModel():pduType is PDU_TYPE_ROUTER_KEY,  protocolVersion must be 1 ", pduType, protocolVersion)
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_UNEXPECTED_PROTOCOL_VERSION, nil, nil),
			protocolVersion, pduType, errors.New("pduType is ROUTER KEY,  protocolVersion must be 1")
	}
	belogs.Debug("parseToPduModel():protocolVersion is ", protocolVersion, "  pduType is ", pduType)
	return nil, protocolVersion, pduType, nil
}

func parseToRtrPduModel(recvByte []byte) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	if len(recvByte) < rtrmodel.PDU_TYPE_MIN_LEN {
		belogs.Error("parseToRtrPduModel(): recv byte's length is too small: ", len(recvByte))
		return rtrmodel.NewRtrErrorReportModel(rtrmodel.PROTOCOL_VERSION_0, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA, nil, nil),
			errors.New(" recv bytes' length is too small")
	}
	buf := bytes.NewReader(recvByte)
	rtrErrorReportPduModel, protocolVersion, pduType, err := parseProtocolVersionAndPduType(buf)
	if err != nil {
		return rtrErrorReportPduModel, err
	}

	switch pduType {
	case rtrmodel.PDU_TYPE_SERIAL_QUERY:

		sq, err := ParseToSerialQuery(buf, protocolVersion)
		if err != nil {
			return rtrPduModel, err
		}
		belogs.Debug("parseToRtrPduModel():  ParseSerialQuery: ", jsonutil.MarshalJson(sq))
		return sq, nil

	case rtrmodel.PDU_TYPE_RESET_QUERY:
		rq, err := ParseToResetQuery(buf, protocolVersion)
		if err != nil {
			return rtrPduModel, err
		}
		belogs.Debug("parseToRtrPduModel():ParseResetQuery: ", jsonutil.MarshalJson(rq))
		return rq, nil

	default:
		belogs.Error("parseToRtrPduModel():pdutype should not recevie by rtr server: ")
		return rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE, nil, nil),
			errors.New("pdutype should not recevie by rtr server, is " + strconv.Itoa(int(pduType)))
	}

}

func processRtrPduModel(rtrPduModel rtrmodel.RtrPduModel) (rtrResponse []rtrmodel.RtrPduModel, err error) {

	pduType := rtrPduModel.GetPduType()
	belogs.Debug("processRtrPduModel():pduType: ", pduType)
	switch pduType {
	case rtrmodel.PDU_TYPE_SERIAL_QUERY:
		serialResponse, err := ProcessSerialQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): ProcessSerialQuery fail: ", err)
			return rtrResponse, err
		}
		belogs.Debug("processRtrPduModel():serialResponse: ", jsonutil.MarshalJson(serialResponse))
		return serialResponse, nil
	case rtrmodel.PDU_TYPE_RESET_QUERY:
		resetResponse, err := ProcessResetQuery(rtrPduModel)
		if err != nil {
			belogs.Error("processRtrPduModel(): ProcessResetQuery fail: ", err)
			return rtrResponse, err
		}
		belogs.Debug("processRtrPduModel():resetResponse: ", jsonutil.MarshalJson(resetResponse))
		return resetResponse, nil
	default:
		return rtrResponse, errors.New("pdutype should not recevie by rtr server, is " + strconv.Itoa(int(pduType)))
	}
}

func sendResponses(conn net.Conn, rtrPduModelResponses *[]rtrmodel.RtrPduModel) (err error) {

	for _, one := range *rtrPduModelResponses {
		belogs.Debug("sendResponses(): send by conn :\r\n", one.PrintBytes())
		_, err = conn.Write(one.Bytes())
		if err != nil {
			belogs.Debug("sendResponses():  conn.Write() fail,  ", jsonutil.MarshalJson(one), err)
			return err
		}
	}
	return nil
}
