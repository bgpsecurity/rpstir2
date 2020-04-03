package rsynchttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"rsync/rsync"
)

// start to rsync
func RsyncStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("rsyncStart(): start")

	go rsync.Start()

	w.WriteJson(httpserver.GetOkHttpResponse())
}
