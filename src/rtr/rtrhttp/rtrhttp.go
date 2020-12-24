package rtrhttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"rtr/rtrtcpclient"
	"rtr/rtrtcpserver"
)

// server send notify to client
func ServerSendSerialNotify(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ServerSendSerialNotify(): start")

	err := rtrtcpserver.SendSerialNotify()
	if err != nil {
		belogs.Error("ServerSendSerialNotify(): SendSerialNotify: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	w.WriteJson(httpserver.GetOkHttpResponse())
}

// start client
func ClientStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ClientStart(): start")

	go rtrtcpclient.Start()
	w.WriteJson(httpserver.GetOkHttpResponse())
}

// stop client
func ClientStop(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ClientEnd(): start")

	go rtrtcpclient.Stop()
	w.WriteJson(httpserver.GetOkHttpResponse())
}

// client send serial query to server
func ClientSendSerialQuery(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ClientSendSerialQuery(): start")

	err := rtrtcpclient.SendSerialQuery()
	if err != nil {
		belogs.Error("ClientSendSerialQuery(): SendSerialQuery: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	w.WriteJson(httpserver.GetOkHttpResponse())
}

// client send reset query to server
func ClientSendResetQuery(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ClientSendResetQuery(): start")

	err := rtrtcpclient.SendResetQuery()
	if err != nil {
		belogs.Error("ClientSendResetQuery(): SendResetQuery: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	w.WriteJson(httpserver.GetOkHttpResponse())
}
