package main

import (
	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/tcpudputil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"rtr/rtrtcp"
)

func main() {
	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("main(): start InitMySql failed:", err)
		return
	}
	defer xormdb.XormEngine.Close()

	// check and init sesessionId
	//go rtr.CheckAndInitSessionId()

	// start server
	belogs.Debug("main(): startTcpServer")
	startTcpServer()

	// block the main thread, to sleep
	select {}

}

func startTcpServer() {

	tcpport := conf.String("rtr::tcpport")
	belogs.Debug("startTcpServer(): tcpport:", tcpport)

	tcpudputil.CreateTcpServer("0.0.0.0:"+tcpport, rtrtcp.RtrServerProcess)
}
