package rtrhttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	httpserver "github.com/cpusoft/goutil/httpserver"

	rtr "rtrproducer/rtr"
)

// start to update db
func RtrUpdate(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("RtrUpdate(): start")
	//check serviceState
	httpclient.Post("https", conf.String("rpstir2::serverHost"), conf.Int("rpstir2::serverHttpsPort"),
		"/sys/servicestate", `{"operate":"enter","state":"rtr"}`)

	go func() {
		nextStep, err := rtr.RtrUpdate()
		belogs.Debug("RtrUpdate():  RtrUpdate end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("RtrUpdate():  RtrUpdate fail", err)
			httpclient.Post("https", conf.String("rpstir2::serverHost"), conf.Int("rpstir2::serverHttpsPort"),
				"/sys/servicestate", `{"operate":"leave","state":"end"}`)
		} else {
			// leave serviceState
			httpclient.Post("https", conf.String("rpstir2::serverHost"), conf.Int("rpstir2::serverHttpsPort"),
				"/sys/servicestate", `{"operate":"leave","state":"rtr"}`)

			// call serial notify to rtr client
			go httpclient.Post("https", conf.String("rpstir2::serverHost"), conf.Int("rpstir2::serverHttpsPort"),
				"/rtr/server/sendserialnotify", "")

			belogs.Info("RtrUpdate():  RtrUpdate end,  nextStep is :", nextStep)
		}
	}()

	w.WriteJson(httpserver.GetOkHttpResponse())
}
