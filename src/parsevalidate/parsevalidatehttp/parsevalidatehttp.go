package parsevalidatehttp

import (
	"errors"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpserver"
	"github.com/cpusoft/goutil/jsonutil"

	"model"
	"parsevalidate/parsevalidate"
)

func ParseValidateStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("ParseValidateStart(): start: ")
	go parsevalidate.ParseValidateStart()

	w.WriteJson(httpserver.GetOkHttpResponse())

}

// upload file to parse;
// only one file
func ParseValidateFile(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("ParseValidateFile(): start: tmpdir:", conf.String("parse::tmpdir"))

	receiveFiles, err := httpserver.ReceiveFiles(conf.String("parse::tmpdir"), req)
	defer httpserver.RemoveReceiveFiles(receiveFiles)
	var certType string
	var certModel interface{}
	var stateModel model.StateModel
	if err == nil {
		if len(receiveFiles) > 0 {
			for _, receiveFile := range receiveFiles {
				certType, certModel, stateModel, err = parsevalidate.ParseValidateFile(receiveFile)
				stateModel.JudgeState()
				belogs.Info("ParseValidateFile(): certType: ", certType,
					"     certModel:", certModel,
					"     stateModel:", stateModel)
				break
			}
		} else {
			err = errors.New("receiveFiles is empty")
		}
	}
	if err != nil {
		belogs.Error("ParseValidateFiles(): ParseValidateFile: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}

	parseCertResponse := model.ParseCertResponse{
		HttpResponse: httpserver.GetOkHttpResponse(),
		CertType:     certType,
		CertModel:    certModel,
		StateModel:   stateModel,
	}
	w.WriteJson(parseCertResponse)
}

// upload file to parse;
// only one file
func ParseFile(w rest.ResponseWriter, req *rest.Request) {
	start := time.Now()
	belogs.Debug("ParseFile(): start: tmpdir:", conf.String("parse::tmpdir"))

	receiveFiles, err := httpserver.ReceiveFiles(conf.String("parse::tmpdir"), req)
	defer httpserver.RemoveReceiveFiles(receiveFiles)

	var certModel interface{}
	if err == nil {
		if len(receiveFiles) > 0 {
			for _, receiveFile := range receiveFiles {
				certModel, err = parsevalidate.ParseFile(receiveFile)
				belogs.Info("ParseValidateFile():receiveFile, certModel:", receiveFile, certModel)
				break
			}
		} else {
			err = errors.New("receiveFiles is empty")
		}
	}
	if err != nil {
		belogs.Error("ParseValidateFiles(): ParseValidateFile: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}
	s := jsonutil.MarshallJsonIndent(certModel)
	belogs.Info("ParseFile(): certModel:", s, "  time(s):", time.Now().Sub(start).Seconds())
	w.WriteJsonString(s)
}

// upload file to parse to get ca repo
func ParseFileSimple(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("ParseFileSimple(): start: tmpdir:", conf.String("parse::tmpdir"))
	receiveFiles, err := httpserver.ReceiveFiles(conf.String("parse::tmpdir"), req)
	defer httpserver.RemoveReceiveFiles(receiveFiles)

	var parseCerSimple model.ParseCerSimple
	if err == nil {
		if len(receiveFiles) > 0 {
			for _, receiveFile := range receiveFiles {
				parseCerSimple, err = parsevalidate.ParseCerSimple(receiveFile)
				break
			}
		} else {
			err = errors.New("receiveFiles is empty")
		}
	}
	if err != nil {
		belogs.Error("ParseFileSimple(): ParseCerSimple: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}

	parseCerSimpleResponse := model.ParseCerSimpleResponse{
		HttpResponse:   httpserver.GetOkHttpResponse(),
		ParseCerSimple: parseCerSimple,
	}
	w.WriteJson(parseCerSimpleResponse)
}
