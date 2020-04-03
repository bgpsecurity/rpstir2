package statistichttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"statistic/statistic"
)

// start to statistic
func StatisticStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("StatisticStart(): start")

	go statistic.Start()

	w.WriteJson(httpserver.GetOkHttpResponse())
}
