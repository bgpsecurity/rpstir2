package rtrtcpclient

import (
	"errors"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/logs"
	ts "github.com/cpusoft/goutil/tcpserver"
)

var RtrTcpClient *ts.TcpClient

func Start() {
	tcpserver := conf.String("rtr::tcpserver")
	tcpport := conf.String("rtr::tcpport")
	rtrTcpClientProcessFunc := new(RtrTcpClientProcessFunc)
	belogs.Info("Start():Rtr Tcp Client: connect to tcpserver:", tcpserver, "    tcpport:", tcpport)

	//CreateTcpClient("127.0.0.1:9999", ClientProcess1)
	RtrTcpClient = ts.NewTcpClient(rtrTcpClientProcessFunc)
	belogs.Debug("Start(): Tcp Client, will start RtrTcpClient %p ", RtrTcpClient)
	go RtrTcpClient.Start(tcpserver + ":" + tcpport)
	belogs.Debug("Start(): Tcp Client, after start RtrTcpClient %p ", RtrTcpClient)

}
func Stop() {
	if RtrTcpClient == nil {
		belogs.Error("Stop():RtrTcpClient is nil fail, should call Start() first ")
		return
	}

	belogs.Info("SendSerialQuery():client, CallStop:", RtrTcpClient)
	RtrTcpClient.CallStop()
}
func SendSerialQuery() (err error) {
	if RtrTcpClient == nil {
		belogs.Error("SendSerialQuery():RtrTcpClient is nil fail, should call Start() first ")
		return errors.New("RtrTcpClient is nil, should call Start() first")
	}

	belogs.Info("SendSerialQuery():client, CallProcessFunc serialquery:", RtrTcpClient)
	RtrTcpClient.CallProcessFunc("serialquery")
	return nil
}

func SendResetQuery() (err error) {
	if RtrTcpClient == nil {
		belogs.Error("SendResetQuery():RtrTcpClient is nil fail, should call Start() first ")
		return errors.New("RtrTcpClient is nil, should call Start() first")
	}

	belogs.Info("SendSerialQuery():client, CallProcessFunc resetquery:", RtrTcpClient)
	RtrTcpClient.CallProcessFunc("resetquery")
	return nil
}
