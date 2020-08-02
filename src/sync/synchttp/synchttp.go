package synchttp

import (
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"model"
	"sync/sync"
)

// start to sync
func SyncStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("SyncStart(): start")

	syncStyle := model.SyncStyle{}
	err := req.DecodeJsonPayload(&syncStyle)
	if err != nil {
		belogs.Error("SyncStart(): DecodeJsonPayload:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	if syncStyle.SyncStyle != "sync" && syncStyle.SyncStyle != "rrdp" && syncStyle.SyncStyle != "rsync" {
		belogs.Error("SyncStart(): syncStyle should be sync or rrdp or rsyncc, it is ", syncStyle.SyncStyle)
		w.WriteJson(httpserver.GetFailHttpResponse(errors.New("SyncStyle should be sync or rrdp or rsync")))
		return
	}
	belogs.Debug("SyncStart(): syncStyle:", syncStyle)

	go sync.Start(syncStyle)
	w.WriteJson(httpserver.GetOkHttpResponse())
}

// get result from rsync
func RsyncResult(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("RsyncResult(): start")

	r := model.SyncResult{}
	err := req.DecodeJsonPayload(&r)
	if err != nil {
		belogs.Error("RsyncResult(): DecodeJsonPayload:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	sync.RsyncResult(&r)
	w.WriteJson(httpserver.GetOkHttpResponse())
}

// get result from rrdp
func RrdpResult(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("RrdpResult(): start")

	r := model.SyncResult{}
	err := req.DecodeJsonPayload(&r)
	if err != nil {
		belogs.Error("RrdpResult(): DecodeJsonPayload:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	sync.RrdpResult(&r)
	w.WriteJson(httpserver.GetOkHttpResponse())
}
