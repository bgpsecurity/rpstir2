package rtrhttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"rtr/rtr"
)

// start to rsync
func RtrUpdate(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("RtrUpdate(): start")

	go rtr.RtrUpdate()

	w.WriteJson(httpserver.GetOkHttpResponse())
}
