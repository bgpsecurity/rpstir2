package rtrserver

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	ts "github.com/cpusoft/goutil/tcpserver"
)

var RtrTcpServer *ts.TcpServer

func RtrServerStart(tcpPort string) {
	//tcpport := conf.String("rpstir2-vc::serverTcpPort")
	belogs.Debug("RtrServerStart(): serverTcpPort:", tcpPort)

	rtrTcpServerProcessFunc := new(RtrTcpServerProcessFunc)
	RtrTcpServer = ts.NewTcpServer(rtrTcpServerProcessFunc)
	belogs.Info("RtrServerStart(): start tcp server on :", tcpPort)
	belogs.Debug("RtrServerStart(): will start RtrTcpServer: %p ", RtrTcpServer)
	go RtrTcpServer.Start("0.0.0.0:" + tcpPort)
	belogs.Debug("RtrServerStart(): after start RtrTcpServer: %p ", RtrTcpServer)

}

func SendSerialNotify() (err error) {

	start := time.Now()
	belogs.Debug("SendSerialNotify():server, start, RtrTcpServer: %p ", RtrTcpServer)
	if RtrTcpServer == nil {
		belogs.Error("SendSerialNotify():RtrTcpServer is nil fail, should start first ")
		return errors.New("RtrTcpServer is nil, should start first")
	}

	rtrPduModelResponse, err := ProcessSerialNotify(PDU_PROTOCOL_VERSION_0)
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
