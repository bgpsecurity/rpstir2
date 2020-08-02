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
	"slurm/slurmhttp"
	"sync/synchttp"
	"sys/syshttp"
	"tal/talhttp"
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
		pprofport := conf.String("pprof::pprofport")
		belogs.Info(http.ListenAndServe("localhost:"+pprofport, nil))
	}()
}

// start server
func startServer() {
	belogs.Info("startServer(): start server, runtime.NumCPU():", runtime.NumCPU())

	runtime.GOMAXPROCS(2 * runtime.NumCPU())

	routes := make([]*rest.Route, 0)

	// tal
	routes = append(routes, rest.Post("/tal/gettals", talhttp.GetTals))
	//sync
	routes = append(routes, rest.Post("/sync/start", synchttp.SyncStart))
	routes = append(routes, rest.Post("/sync/rrdpresult", synchttp.RrdpResult))
	routes = append(routes, rest.Post("/sync/rsyncresult", synchttp.RsyncResult))
	// rrdp(delta)
	routes = append(routes, rest.Post("/rrdp/start", rrdphttp.RrdpStart))
	// rsync
	routes = append(routes, rest.Post("/rsync/start", rsynchttp.RsyncStart))

	// parsevalidate
	routes = append(routes, rest.Post("/parsevalidate/start", parsevalidatehttp.ParseValidateStart))
	routes = append(routes, rest.Post("/parsevalidate/file", parsevalidatehttp.ParseValidateFile))
	routes = append(routes, rest.Post("/parsevalidate/parsefile", parsevalidatehttp.ParseFile))
	routes = append(routes, rest.Post("/parsevalidate/parsefilesimple", parsevalidatehttp.ParseFileSimple))

	// chainvalidate
	routes = append(routes, rest.Post("/chainvalidate/start", chainvalidatehttp.ChainValidateStart))

	// sys
	routes = append(routes, rest.Post("/sys/initreset", syshttp.InitReset))
	routes = append(routes, rest.Post("/sys/detailstates", syshttp.DetailStates))
	routes = append(routes, rest.Post("/sys/summarystates", syshttp.SummaryStates))
	routes = append(routes, rest.Post("/sys/results", syshttp.Results))
	routes = append(routes, rest.Post("/sys/exportroas", syshttp.ExportRoas))

	// slurm
	routes = append(routes, rest.Post("/slurm/upload", slurmhttp.SlurmUpload))
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
