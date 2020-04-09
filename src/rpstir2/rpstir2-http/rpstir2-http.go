package main

import (
	"runtime"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpserver"
	_ "github.com/cpusoft/goutil/logs"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"net/http"
	_ "net/http/pprof"

	"chainvalidate/chainvalidatehttp"
	"parsevalidate/parsevalidatehttp"
	"rrdp/rrdphttp"
	"rsync/rsynchttp"
	"rtr/rtrhttp"
	"sys/syshttp"
)

func main() {

	startPprof()

	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("main(): start InitMySql failed:", err)
		return
	}

	defer xormdb.XormEngine.Close()

	// start server
	startServer()
	// block the main thread, to sleep
	select {}
}

func startPprof() {
	go func() {
		belogs.Info(http.ListenAndServe("localhost:8084", nil))
	}()
}

// start server
func startServer() {
	belogs.Info("startServer(): start server, runtime.NumCPU():", runtime.NumCPU())

	runtime.GOMAXPROCS(2 * runtime.NumCPU())

	routes := make([]*rest.Route, 0)

	// rsync
	routes = append(routes, rest.Post("/rsync/start", rsynchttp.RsyncStart))
	// rrdp(delta)
	routes = append(routes, rest.Post("/rrdp/start", rrdphttp.RrdpStart))

	// parsevalidate
	routes = append(routes, rest.Post("/parsevalidate/start", parsevalidatehttp.ParseValidateStart))
	routes = append(routes, rest.Post("/parsevalidate/file", parsevalidatehttp.ParseValidateFile))
	routes = append(routes, rest.Post("/parsevalidate/filerepo", parsevalidatehttp.ParseValidateFileRepo))

	// chainvalidate
	routes = append(routes, rest.Post("/chainvalidate/start", chainvalidatehttp.ChainValidateStart))

	// rtr
	routes = append(routes, rest.Post("/rtr/update", rtrhttp.RtrUpdate))

	// sys
	routes = append(routes, rest.Post("/sys/init", syshttp.Init))
	routes = append(routes, rest.Post("/sys/reset", syshttp.Reset))
	routes = append(routes, rest.Post("/sys/detailstates", syshttp.DetailStates))
	routes = append(routes, rest.Post("/sys/summarystates", syshttp.SummaryStates))
	routes = append(routes, rest.Post("/sys/results", syshttp.Results))

	// make router
	router, err := rest.MakeRouter(
		routes...,
	)

	if err != nil {
		belogs.Error("startServer(): failed: err:", err)
		return
	}
	// if have http port, then sart http server, default is off
	httpport := conf.String("rpstir2::httpport")
	if httpport != "" {
		go func() {
			belogs.Info("startServer():start http on: ", ":"+httpport)
			httpserver.ListenAndServe(":"+httpport, &router)
		}()
	}

}
