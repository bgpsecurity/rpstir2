package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpserver"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/osutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"chainvalidate/chainvalidatehttp"
	"parsevalidate/parsevalidatehttp"
	"rrdp/rrdphttp"
	"rsync/rsynchttp"
	rtrhttp "rtr/rtrhttp"
	"rtr/rtrtcpserver"
	rtrproducerhttp "rtrproducer/rtrhttp"
	"sync/synchttp"
	"sys/syshttp"
	"tal/talhttp"
)

func main() {

	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("main(): start InitMySql failed:", err)
		fmt.Println("rpstir2 failed to start when connecting to MySQL, the error is ", err)
		return
	}

	defer xormdb.XormEngine.Close()

	// start server
	startPprof()
	startTcpServer()
	startServer()
	// block the main thread, to sleep
	select {}
}

func startPprof() {
	go func() {
		pprofport := conf.String("rpstir2::pprofHttpPort")
		belogs.Info(http.ListenAndServe(":"+pprofport, nil))
	}()
}
func startTcpServer() {
	belogs.Debug("startTcpServer():will start tcp server")
	rtrtcpserver.Start()
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
	routes = append(routes, rest.Post("/sys/servicestate", syshttp.ServiceState))
	routes = append(routes, rest.Post("/sys/results", syshttp.Results))
	routes = append(routes, rest.Post("/sys/exportroas", syshttp.ExportRoas))

	// rtr Update
	routes = append(routes, rest.Post("/rtrproducer/update", rtrproducerhttp.RtrUpdate))

	// rtr
	routes = append(routes, rest.Post("/rtr/server/sendserialnotify", rtrhttp.ServerSendSerialNotify))
	routes = append(routes, rest.Post("/rtr/client/start", rtrhttp.ClientStart))
	routes = append(routes, rest.Post("/rtr/client/stop", rtrhttp.ClientStop))
	routes = append(routes, rest.Post("/rtr/client/sendserialquery", rtrhttp.ClientSendSerialQuery))
	routes = append(routes, rest.Post("/rtr/client/sendresetquery", rtrhttp.ClientSendResetQuery))

	/////////////////////

	// make router
	router, err := rest.MakeRouter(
		routes...,
	)

	if err != nil {
		belogs.Error("startServer(): failed: err:", err)
		return
	}
	// if have http port, then sart http server, default is off
	httpport := conf.String("rpstir2::serverHttpPort")
	if httpport != "" {
		go func() {
			belogs.Info("startServer():start http on: ", ":"+httpport)
			httpserver.ListenAndServe(":"+httpport, &router)
		}()
	}

	/////////////////////
	// Advanced functions
	/////////////////////
	// if have https port, then start https server, default is on
	httpsport := conf.String("rpstir2::serverHttpsPort")
	if httpsport != "" {
		go func() {
			belogs.Info("startServer():start https on: ", ":"+httpsport)
			confPath, _ := osutil.GetCurrentOrParentAbsolutePath("conf")
			path := confPath + string(os.PathSeparator) + "cert" +
				string(os.PathSeparator) + "rpstir2" +
				string(os.PathSeparator)
			crtFile := path + conf.String("rpstir2::serverCrt")
			keyFile := path + conf.String("rpstir2::serverKey")
			httpserver.ListenAndServeTLS(":"+httpsport, crtFile, keyFile, &router)
		}()
	}
	/////////////////////
}
