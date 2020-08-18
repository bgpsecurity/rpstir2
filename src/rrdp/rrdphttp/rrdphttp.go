package rrdphttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"model"
	"rrdp/rrdp"
)

// start to rrdp from sync
func RrdpStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("RrdpStart(): start")

	syncUrls := model.SyncUrls{}
	err := req.DecodeJsonPayload(&syncUrls)
	belogs.Debug("RrdpStart(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RrdpStart(): DecodeJsonPayload:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}

	go rrdp.Start(&syncUrls)

	w.WriteJson(httpserver.GetOkHttpResponse())
}
