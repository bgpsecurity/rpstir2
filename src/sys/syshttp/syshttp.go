package syshttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	httpserver "github.com/cpusoft/goutil/httpserver"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	"sys/sys"
)

//
func Reset(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("Reset()")

	sys.InitReset(false)

	w.WriteJson(httpserver.GetOkHttpResponse())
}
func Init(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("Init()")

	sys.InitReset(true)

	w.WriteJson(httpserver.GetOkHttpResponse())
}

// detail
func DetailStates(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("DetailStates()")

	detailStates, err := sys.DetailStates()
	if err != nil {
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	belogs.Info("DetailStates():detailStates:", jsonutil.MarshalJson(detailStates))

	stateResponse := model.StateResponse{
		HttpResponse: httpserver.GetOkHttpResponse(),
		State:        detailStates,
	}
	w.WriteJson(stateResponse)
}

// summary
func SummaryStates(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("SummaryStates()")

	summaryStates, err := sys.SummaryStates()
	if err != nil {
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	belogs.Info("SummaryStates():summaryStates:", jsonutil.MarshalJson(summaryStates))

	stateResponse := model.StateResponse{
		HttpResponse: httpserver.GetOkHttpResponse(),
		State:        summaryStates,
	}
	w.WriteJson(stateResponse)
}

// just return valid/warning/invalid count in cer/roa/mft/crl
func Results(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("Results()")

	results, err := sys.Results()
	if err != nil {
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	belogs.Info("Results():results:", jsonutil.MarshalJson(results))
	w.WriteJson(results)
}
