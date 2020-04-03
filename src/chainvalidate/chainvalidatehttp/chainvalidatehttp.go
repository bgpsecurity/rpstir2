package chainvalidatehttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	"github.com/cpusoft/goutil/httpserver"

	"chainvalidate/chainvalidate"
)

// upload file to parse
func ChainValidateStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ChainValidateStart(): start")

	go chainvalidate.ChainValidateStart()

	w.WriteJson(httpserver.GetOkHttpResponse())

}
