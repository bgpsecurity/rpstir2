package rtrtcpserver

import (
	"bytes"
	"net"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
	rtrtcp "rtr/rtrtcp"
)

type RtrTcpServerProcessFunc struct {
}

func (rs *RtrTcpServerProcessFunc) OnConnect(conn *net.TCPConn) {

}
func (rs *RtrTcpServerProcessFunc) OnReceiveAndSend(conn *net.TCPConn, receiveData []byte) (err error) {

	start := time.Now()
	buf := bytes.NewReader(receiveData)
	// parse []byte --> rtrpdumodel
	rtrPduModel, err := rtrtcp.ParseToRtrPduModel(buf)
	if err != nil {
		belogs.Error("OnReceiveAndSend():server,  ParseToRtrPduModel fail: ", convert.PrintBytes(receiveData, 8), err)
		err = rtrtcp.SendErrorResponse(conn, err)
		if err != nil {
			belogs.Error("OnReceiveAndSend():server, SendErrorResponse fail: ", err)
		}
		return err
	}
	belogs.Info("OnReceiveAndSend():server get rtrPduModel:", jsonutil.MarshalJson(rtrPduModel))

	rtrPduModelResponses := make([]rtrmodel.RtrPduModel, 0)
	// process rtrpdumodel --> response rtrpdumodels
	rtrPduModelResponses, err = rtrtcp.ProcessRtrPduModel(buf, rtrPduModel)
	if err != nil {
		belogs.Error("OnReceiveAndSend():server,  processRtrPduModel fail: ", jsonutil.MarshalJson(rtrPduModel), err)
		err = rtrtcp.SendErrorResponse(conn, err)
		if err != nil {
			belogs.Error("OnReceiveAndSend():server, SendErrorResponse fail: ", err)
		}
		return err
	}
	belogs.Info("OnReceiveAndSend():server process rtrPduModel, and assemable responses, len(responses) is ", len(rtrPduModelResponses))

	// send response rtrpdumodels
	err = rtrtcp.SendResponses(conn, rtrPduModelResponses)
	if err != nil {
		belogs.Error("OnReceiveAndSend():server, sendResponses fail: ", err)
		// send internal error
		return err
	}
	belogs.Info("OnReceiveAndSend(): server send responses ok, len(responses) is ", len(rtrPduModelResponses),
		" ,  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
func (rs *RtrTcpServerProcessFunc) OnClose(conn *net.TCPConn) {

}
func (rs *RtrTcpServerProcessFunc) ActiveSend(conn *net.TCPConn, sendData []byte) (err error) {
	belogs.Debug("ActiveSend():len(sendData):", len(sendData))
	start := time.Now()
	conn.SetWriteBuffer(len(sendData))
	n, err := conn.Write(sendData)
	if err != nil {
		belogs.Debug("ActiveSend():server, conn.Write() fail,  ", convert.Bytes2String(sendData), err)
		return err
	}
	belogs.Info("ActiveSend(): conn.Write() ok, len(sendData), n:", len(sendData), n, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
