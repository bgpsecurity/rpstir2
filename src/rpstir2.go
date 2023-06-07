package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	chainvalidate "rpstir2-chainvalidate"
	clear "rpstir2-clear"
	parsevalidate "rpstir2-parsevalidate"
	rtrclient "rpstir2-rtrclient"
	rtrproducer "rpstir2-rtrproducer"
	rtrserver "rpstir2-rtrserver"
	entiremixsync "rpstir2-sync-entire/mixsync"
	entirerrdp "rpstir2-sync-entire/rrdp"
	entirersync "rpstir2-sync-entire/rsync"
	entiresync "rpstir2-sync-entire/sync"
	tal "rpstir2-sync-tal"
	sys "rpstir2-sys"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/xormdb"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func main() {

	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("main(): start InitMySql failed:", err)
		fmt.Println("rpstir2 failed to start, ", err)
		return
	}
	defer xormdb.XormEngine.Close()
	// start rp server
	go startRpServer()

	go startVcServer()
	go startTcpServer()

	select {}
}

// start server
func startRpServer() {
	start := time.Now()
	var g errgroup.Group

	serverHttpPort := conf.String("rpstir2-rp::serverHttpPort")
	serverHttpsPort := conf.String("rpstir2-rp::serverHttpsPort")
	serverCrt := conf.String("rpstir2-rp::serverCrt")
	serverKey := conf.String("rpstir2-rp::serverKey")
	belogs.Info("startRpServer(): start server, serverHttpPort:", serverHttpPort,
		"    serverHttpsPort:", serverHttpsPort, "   serverCrt:", serverCrt,
		"    serverKey:", serverKey)

	//gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.POST("/tal/gettals", tal.GetTals)

	engine.POST("/entiresync/syncstart", entiremixsync.SyncStart)
	//engine.POST("/entiresync/syncstart", entiresync.SyncStart)
	engine.POST("/entiresync/rrdpresult", entiresync.RrdpResult)
	engine.POST("/entiresync/rsyncresult", entiresync.RsyncResult)
	engine.POST("/entiresync/rrdprequest", entirerrdp.RrdpRequest)
	engine.POST("/entiresync/rsyncrequest", entirersync.RsyncRequest)
	engine.POST("/parsevalidate/start", parsevalidate.ParseValidateStart)
	engine.POST("/parsevalidate/file", parsevalidate.ParseValidateFile)
	engine.POST("/parsevalidate/parsefile", parsevalidate.ParseFile)
	engine.POST("/parsevalidate/parsefilesimple", parsevalidate.ParseFileSimple)
	engine.POST("/chainvalidate/start", chainvalidate.ChainValidateStart)
	engine.POST("/clear/start", clear.ClearStart)
	engine.POST("/sys/initreset", sys.InitReset)
	engine.POST("/sys/servicestate", sys.ServiceState)
	engine.POST("/sys/results", sys.Results)
	engine.POST("/sys/exportroas", sys.ExportRoas)

	/////////////////////

	if serverHttpPort != "" {
		belogs.Info("startRpServer(): http on :", serverHttpPort)
		g.Go(func() error {
			belogs.Info("startRpServer(): server run http on :", serverHttpPort)
			err := engine.Run(":" + serverHttpPort)
			if err != nil {
				belogs.Error("startRpServer(): http fail, will exit, err:", serverHttpPort, err)
			}
			return err
		})
	}

	if serverHttpsPort != "" {
		belogs.Info("startRpServer(): https on :", serverHttpsPort)
		g.Go(func() error {
			certsPath := osutil.GetParentPath() + "/conf/cert/"
			belogs.Info("startRpServer(): server run https on :", serverHttpsPort, certsPath+serverCrt, certsPath+serverKey)
			err := engine.RunTLS(":"+serverHttpsPort, certsPath+serverCrt, certsPath+serverKey)
			if err != nil {
				belogs.Error("startRpServer(): https fail, will exit, err:", serverHttpsPort, err)
			}
			return err
		})
	}

	go func() {
		pprofport := conf.String("rpstir2-rp::pprofHttpPort")
		belogs.Info(http.ListenAndServe(":"+pprofport, nil))
	}()

	if err := g.Wait(); err != nil {
		belogs.Error("startRpServer(): fail, will exit, err:", err)
	}
	belogs.Info("startRpServer(): server end, time(s):", time.Now().Sub(start).Seconds())

}

// start vc server
func startVcServer() {
	start := time.Now()
	var g errgroup.Group

	serverHttpPort := conf.String("rpstir2-vc::serverHttpPort")
	serverHttpsPort := conf.String("rpstir2-vc::serverHttpsPort")
	serverCrt := conf.String("rpstir2-vc::serverCrt")
	serverKey := conf.String("rpstir2-vc::serverKey")
	belogs.Info("startVcServer(): start server, serverHttpPort:", serverHttpPort,
		"    serverHttpsPort:", serverHttpsPort, "   serverCrt:", serverCrt,
		"    serverKey:", serverKey)
	//gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.POST("/rtrproducer/updatefromsync", rtrproducer.RtrUpdateFromSync)
	engine.POST("/sys/initreset", sys.InitReset)
	engine.POST("/rtr/server/sendserialnotify", rtrserver.ServerSendSerialNotify)
	engine.POST("/rtr/client/start", rtrclient.ClientStart)
	engine.POST("/rtr/client/stop", rtrclient.ClientStop)
	engine.POST("/rtr/client/sendserialquery", rtrclient.ClientSendSerialQuery)
	engine.POST("/rtr/client/sendresetquery", rtrclient.ClientSendResetQuery)

	/////////////////////

	if serverHttpPort != "" {
		belogs.Info("startVcServer(): http on :", serverHttpsPort)
		g.Go(func() error {
			belogs.Info("startVcServer(): server run http on :", serverHttpPort)
			err := engine.Run(":" + serverHttpPort)
			if err != nil {
				belogs.Error("startVcServer(): http fail, will exit, err:", serverHttpPort, err)
			}
			return err
		})
	}

	if serverHttpsPort != "" {
		belogs.Info("startVcServer(): https on :", serverHttpsPort)
		g.Go(func() error {
			certsPath := osutil.GetParentPath() + "/conf/cert/"
			belogs.Info("startVcServer(): server run https on :", serverHttpsPort, certsPath+serverCrt, certsPath+serverKey)
			err := engine.RunTLS(":"+serverHttpsPort, certsPath+serverCrt, certsPath+serverKey)
			if err != nil {
				belogs.Error("startVcServer(): https fail, will exit, err:", serverHttpsPort, err)
			}
			return err
		})
	}

	go func() {
		pprofport := conf.String("rpstir2-vc::pprofHttpPort")
		belogs.Info(http.ListenAndServe(":"+pprofport, nil))
	}()

	if err := g.Wait(); err != nil {
		belogs.Error("startVcServer(): fail, will exit, err:", err)
	}
	belogs.Info("startVcServer(): server end, time(s):", time.Now().Sub(start).Seconds())

}

func startTcpServer() {
	// rtrtcp
	tcpPort := conf.String("rpstir2-vc::serverTcpPort")
	belogs.Debug("startTcpServer():will start tcp server:", tcpPort)
	rtrserver.RtrServerStart(tcpPort)

}
