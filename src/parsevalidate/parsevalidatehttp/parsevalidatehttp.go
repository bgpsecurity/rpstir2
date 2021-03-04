package parsevalidatehttp

import (
	"errors"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/httpserver"
	"github.com/cpusoft/goutil/jsonutil"

	"model"
	"parsevalidate/parsevalidate"
)

func ParseValidateStart(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("ParseValidateStart(): start: ")

	//check serviceState
	httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"parsevalidate"}`, false)

	go func() {
		nextStep, err := parsevalidate.ParseValidateStart()
		belogs.Debug("ParseValidateStart():  ParseValidateStart end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("ParseValidateStart():  ParseValidateStart fail", err)
			httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {
			httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"parsevalidate"}`, false)
			// will call ChainValidate
			go httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/chainvalidate/start", "", false)
			belogs.Info("ParseValidateStart():  sync.Start end,  nextStep is :", nextStep)
		}

	}()

	w.WriteJson(httpserver.GetOkHttpResponse())

}

// upload file to parse;
// only one file
func ParseValidateFile(w rest.ResponseWriter, req *rest.Request) {
	belogs.Debug("ParseValidateFile(): start: tmpDir:", conf.String("parse::tmpDir"))

	receiveFiles, err := httpserver.ReceiveFiles(conf.String("parse::tmpDir"), req)
	defer httpserver.RemoveReceiveFiles(receiveFiles)
	belogs.Debug("ParseValidateFile(): receiveFiles:", receiveFiles)

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
	belogs.Debug("ParseFile(): start: tmpDir:", conf.String("parse::tmpDir"))

	receiveFiles, err := httpserver.ReceiveFiles(conf.String("parse::tmpDir"), req)
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
	belogs.Debug("ParseFileSimple(): start: tmpDir:", conf.String("parse::tmpDir"))
	receiveFiles, err := httpserver.ReceiveFiles(conf.String("parse::tmpDir"), req)
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
