package synchttp

import (
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	httpserver "github.com/cpusoft/goutil/httpserver"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	"sync/sync"
)

// start to sync
func SyncStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("SyncStart(): start")

	// get syncStyle
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

	//check serviceState
	resp, body, err := httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"sync"}`, false)
	if err != nil {
		belogs.Error("SyncStart(): /sys/servicestate connecteds failed, err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	resp.Body.Close()
	ssr := model.ServiceStateResponse{}
	jsonutil.UnmarshalJson(body, &ssr)
	belogs.Debug("SyncStart(): get /sys/servicestate serviceStateResponse:", jsonutil.MarshalJson(ssr))
	if ssr.Result == "fail" {
		belogs.Error("SyncStart(): /sys/servicestate ssr.Result is fail, err:", ssr)
		w.WriteJson(httpserver.GetFailHttpResponse(errors.New(ssr.Msg)))
		return
	}

	go func() {
		nextStep, err := sync.SyncStart(syncStyle)
		belogs.Debug("SyncStart():  SyncStart end,  nextStep is :", nextStep, err)

		if err != nil {
			// will end this whole sync
			belogs.Error("SyncStart(): SyncStart fail,  syncStyle is :", syncStyle, err)
			httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {

			// will go next step
			if nextStep == "fullsync" {
				// leave current sync now, and start new fullsync
				httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
					"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)

				go httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
					"/sys/initreset", `{"sysStyle":"fullsync", "syncStyle":"`+syncStyle.SyncStyle+`"}`, false)
			} else if nextStep == "parsevalidate" {
				// will end sync ,and will start next step
				httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
					"/sys/servicestate", `{"operate":"leave","state":"sync"}`, false)

				go httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
					"/parsevalidate/start", "", false)
			}
			belogs.Info("SyncStart():  sync.Start end,  nextStep is :", nextStep)
		}

	}()
	w.WriteJson(httpserver.GetOkMsgHttpResponse("The synchronization and validation processes will run in the background."))
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

// sync from local, for history repo data: need reset and just start from diff, then parse....
func SyncLocalStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("SyncLocalStart(): start")
	go sync.LocalStart()
	w.WriteJson(httpserver.GetOkMsgHttpResponse("The validation processes will run in the background."))
}
