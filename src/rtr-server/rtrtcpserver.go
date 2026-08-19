package rtrserver

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	ts "github.com/cpusoft/goutil/tcpserver"
)

var RtrTcpServer *ts.TcpServer

func RtrServerStart(tcpPort string) {
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
	rtrProtocolVersion := uint8(conf.Int("rtr::protocolVersion"))
	belogs.Debug("SendSerialNotify() rtrProtocolVersion:", rtrProtocolVersion)

	rtrPduModelResponse, err := processSerialNotify(rtrProtocolVersion)
	if err != nil {
		belogs.Error("SendSerialNotify():server, processSerialNotify fail: ", err)
		return err
	}
	belogs.Debug("SendSerialNotify():server, RtrTcpServer:", RtrTcpServer, " processSerialNotify rtrPduModelResponse: ", jsonutil.MarshalJson(rtrPduModelResponse))

	// send response rtrpdumodels
	err = RtrTcpServer.Broadcast(rtrPduModelResponse.Bytes())
	if err != nil {
		belogs.Error("SendSerialNotify():Broadcast fail", "err", err)
		return err
	}
	belogs.Info("SendSerialNotify(): Broadcast ok, rtrPduModelResponse:", jsonutil.MarshalJson(rtrPduModelResponse), "   time(s):", time.Since(start))
	return nil

}
