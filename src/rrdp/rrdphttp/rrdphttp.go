package rrdphttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"rrdp/rrdp"
)

// start to rrdp
func RrdpStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("RrdpStart(): start")

	go rrdp.Start()

	w.WriteJson(httpserver.GetOkHttpResponse())
}
