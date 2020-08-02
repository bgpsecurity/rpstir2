package talhttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	"tal/tal"
)

//
func GetTals(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("GetTals")

	talModels, err := tal.GetTals()
	belogs.Debug("GetTals(): GetTals, talModels:", jsonutil.MarshalJson(talModels))
	if err != nil {
		w.WriteJson(httpserver.GetFailHttpResponse(err))
	} else {
		talResponse := model.TalResponse{
			HttpResponse: httpserver.GetOkHttpResponse(),
			TalModels:    talModels,
		}
		w.WriteJson(talResponse)
	}
}
