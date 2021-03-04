package chainvalidatehttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/httpserver"

	"chainvalidate/chainvalidate"
)

// upload file to parse
func ChainValidateStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("ChainValidateStart(): start")

	//check serviceState
	httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"chainvalidate"}`, false)

	go func() {
		nextStep, err := chainvalidate.ChainValidateStart()
		belogs.Debug("ChainValidateStart():  ChainValidateStart end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("ParseValidateStart():  ChainValidateStart fail", err)
			httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {
			// leave serviceState
			httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"chainvalidate"}`, false)

			go httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/rtrproducer/update", "", false)
			belogs.Info("ParseValidateStart():  sync.Start end,  nextStep is :", nextStep)
		}
	}()

	w.WriteJson(httpserver.GetOkHttpResponse())

}
