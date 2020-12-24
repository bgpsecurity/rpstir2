package rtrtcpserver

import (
	"errors"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	ts "github.com/cpusoft/goutil/tcpserver"

	rtrmodel "rtr/model"
	rtrtcp "rtr/rtrtcp"
)

var RtrTcpServer *ts.TcpServer

func Start() {
	tcpport := conf.String("rpstir2::serverTcpPort")
	belogs.Debug("Start(): tcpport:", tcpport)

	rtrTcpServerProcessFunc := new(RtrTcpServerProcessFunc)
	RtrTcpServer = ts.NewTcpServer(rtrTcpServerProcessFunc)
	belogs.Info("startTcpServer(): start tcp server on :", tcpport)
	belogs.Debug("Start(): will start RtrTcpServer: %p ", RtrTcpServer)
	go RtrTcpServer.Start("0.0.0.0:" + tcpport)
	belogs.Debug("Start(): after start RtrTcpServer: %p ", RtrTcpServer)

}

func SendSerialNotify() (err error) {

	start := time.Now()
	belogs.Debug("SendSerialNotify():server, start, RtrTcpServer: %p ", RtrTcpServer)
	if RtrTcpServer == nil {
		belogs.Error("SendSerialNotify():RtrTcpServer is nil fail, should call Start() first ")
		return errors.New("RtrTcpServer is nil, should call Start() first")
	}

	rtrPduModelResponse, err := rtrtcp.ProcessSerialNotify(rtrmodel.PROTOCOL_VERSION_0)
	if err != nil {
		belogs.Error("SendSerialNotify():server, ProcessSerialNotify fail: ", err)
		return err
	}
	belogs.Debug("SendSerialNotify():server, RtrTcpServer:", RtrTcpServer, " ProcessSerialNotify rtrPduModelResponse: ", jsonutil.MarshalJson(rtrPduModelResponse))

	// send response rtrpdumodels
	RtrTcpServer.ActiveSend(rtrPduModelResponse.Bytes())
	belogs.Info("SendSerialNotify(): ok,   rtrPduModelResponse:", jsonutil.MarshalJson(rtrPduModelResponse), "   time(s):", time.Now().Sub(start).Seconds())
	return nil

}
