package rtrclient

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	ts "github.com/cpusoft/goutil/tcpserver"
)

var RtrTcpClient *ts.TcpClient

func clientStart() {
	tcpserver := conf.String("rtr::tcpserver")
	tcpport := conf.String("rtr::tcpport")
	rtrTcpClientProcessFunc := new(RtrTcpClientProcessFunc)
	belogs.Info("clientStart():Rtr Tcp Client: connect to tcpserver:", tcpserver, "    tcpport:", tcpport)

	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	RtrTcpClient = ts.NewTcpClient(rtrTcpClientProcessFunc)
	belogs.Debug("clientStart(): Tcp Client, will start RtrTcpClient %p ", RtrTcpClient)
	go RtrTcpClient.Start(tcpserver + ":" + tcpport)
	belogs.Debug("clientStart(): Tcp Client, after start RtrTcpClient %p ", RtrTcpClient)

}
func clientStop() {
	if RtrTcpClient == nil {
		belogs.Error("clientStop():RtrTcpClient is nil fail, should start first ")
		return
	}

	belogs.Info("clientStop():client, CallStop:", RtrTcpClient)
	RtrTcpClient.CallStop()
}
func clientSendSerialQuery() (err error) {
	if RtrTcpClient == nil {
		belogs.Error("clientSendSerialQuery():RtrTcpClient is nil fail, should start first ")
		return errors.New("RtrTcpClient is nil, should start first")
	}

	belogs.Info("clientSendSerialQuery():client, CallProcessFunc serialquery:", RtrTcpClient)
	RtrTcpClient.CallProcessFunc("serialquery")
	return nil
}

func clientSendResetQuery() (err error) {
	if RtrTcpClient == nil {
		belogs.Error("clientSendResetQuery():RtrTcpClient is nil fail, should start first ")
		return errors.New("RtrTcpClient is nil, should start first")
	}

	belogs.Info("clientSendResetQuery():client, CallProcessFunc resetquery:", RtrTcpClient)
	RtrTcpClient.CallProcessFunc("resetquery")
	return nil
}
