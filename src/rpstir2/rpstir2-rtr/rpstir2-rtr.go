package main

import (
	"runtime"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	_ "github.com/cpusoft/goutil/conf"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpserver"
	_ "github.com/cpusoft/goutil/logs"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"rtr/rtrhttp"
	"rtr/rtrtcpserver"
)

func main() {
	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("main(): start InitMySql failed:", err)
		return
	}
	defer xormdb.XormEngine.Close()

	runtime.GOMAXPROCS(2 * runtime.NumCPU())

	// start server
	startTcpServer()
	startHttpServer()

	// block the main thread, to sleep
	select {}
}

func startTcpServer() {
	belogs.Debug("startTcpServer():will start rtr tcp server")
	rtrtcpserver.Start()
}

// start server
func startHttpServer() {
	belogs.Debug("startHttpServer():will start rtr http server")
	routes := make([]*rest.Route, 0)

	// rtr
	routes = append(routes, rest.Post("/rtr/update", rtrhttp.RtrUpdate))
	routes = append(routes, rest.Post("/rtr/server/sendserialnotify", rtrhttp.ServerSendSerialNotify))
	routes = append(routes, rest.Post("/rtr/client/start", rtrhttp.ClientStart))
	routes = append(routes, rest.Post("/rtr/client/stop", rtrhttp.ClientStop))
	routes = append(routes, rest.Post("/rtr/client/sendserialquery", rtrhttp.ClientSendSerialQuery))
	routes = append(routes, rest.Post("/rtr/client/sendresetquery", rtrhttp.ClientSendResetQuery))

	// make router
	router, err := rest.MakeRouter(
		routes...,
	)

	if err != nil {
		belogs.Error("startHttpServer(): rtr failed: err:", err)
		return
	}
	// if have http port, then sart http server, default is off
	httpport := conf.String("rtr::httpport")
	if httpport != "" {
		go func() {
			belogs.Info("startHttpServer():start rtr http on :", httpport)
			httpserver.ListenAndServe(":"+httpport, &router)
		}()
	}

}
