package rsynchttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"model"
	"rsync/rsync"
)

// start to rsync from sync
func RsyncStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("RsyncStart(): start")

	syncUrls := model.SyncUrls{}
	err := req.DecodeJsonPayload(&syncUrls)
	belogs.Debug("RsyncStart(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RsyncStart(): DecodeJsonPayload:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}

	go rsync.Start(&syncUrls)

	w.WriteJson(httpserver.GetOkHttpResponse())
}

// start local rsync
func RsyncLocalStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("RsyncLocalStart(): start")

	syncUrls := model.SyncUrls{}
	err := req.DecodeJsonPayload(&syncUrls)
	belogs.Debug("RsyncStart(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RsyncStart(): DecodeJsonPayload:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}

	rsyncResult, err := rsync.LocalStart(&syncUrls)
	if err != nil {
		belogs.Error("RsyncLocalStart(): LocalStart:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	belogs.Debug("RsyncLocalStart(): rsyncResult:", rsyncResult)

	rsyncResultResponse := model.RsyncResultResponse{
		HttpResponse: httpserver.GetOkHttpResponse(),
		RsyncResult:  rsyncResult,
	}
	w.WriteJson(rsyncResultResponse)

}
