package slurmhttp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	conf "github.com/cpusoft/goutil/conf"
	httpserver "github.com/cpusoft/goutil/httpserver"

	"slurm/slurm"
)

// upload file to parse
func SlurmUpload(w rest.ResponseWriter, req *rest.Request) {
	belogs.Info("SlurmUpload(): start: slurmpath:", conf.VariableString("slurm::slurmpath"))

	receiveFiles, err := httpserver.ReceiveFiles(conf.VariableString("slurm::slurmpath"), req)
	if err == nil {
		err = slurm.UploadFiles(receiveFiles)
	}
	if err != nil {
		belogs.Error("SlurmUpload(): ReceiveFiles: err:", err)
		w.WriteJson(httpserver.GetFailHttpResponse(err))
		return
	}

	w.WriteJson(httpserver.GetOkHttpResponse())
}
