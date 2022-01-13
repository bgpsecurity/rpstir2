package sys

import (
	"errors"
	"net/http"
	"os"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

//
func InitReset(c *gin.Context) {
	belogs.Debug("InitReset()")
	sysStyle := SysStyle{}
	err := c.ShouldBindJSON(&sysStyle)
	if err != nil {
		belogs.Error("InitReset(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("InitReset():get sysStyle:", jsonutil.MarshalJson(sysStyle))
	if sysStyle.SysStyle != "init" && sysStyle.SysStyle != "fullsync" && sysStyle.SysStyle != "resetall" {
		belogs.Error("InitReset(): SysStyle should be init or fullsync or resetall, it is ", sysStyle.SysStyle)
		ginserver.ResponseFail(c, errors.New("SysStyle should be init or fullsync or resetall"), "")
		return
	}
	belogs.Debug("InitReset(): sysStyle:", sysStyle)

	go func() {
		err := initReset(sysStyle)
		if err == nil && sysStyle.SysStyle == "fullsync" {
			url := "https://" + conf.String("rpstir2-rp::serverHost") + ":" +
				conf.String("rpstir2-rp::serverHttpsPort")
			var path string
			if sysStyle.SyncPolicy == "direct" {
				path = url + "/directsync/directurlstart"
				belogs.Info("initReset(): will call direct url first:", path)
				err = httpclient.PostAndUnmarshalResponseModel(path, ``, false, nil)
				if err != nil {
					belogs.Error("InitReset(): call direct url fail,err: ", err)
					ginserver.ResponseFail(c, errors.New("direct url fail"), "")
					return
				}

				path = url + "/directsync/directsyncstart"
				belogs.Info("initReset(): will call direct sync second:", path)
				go httpclient.Post(path, ``, false)

			} else if sysStyle.SyncPolicy == "entire" {
				path = url + "/entiresync/syncstart"
				belogs.Info("initReset(): will call entire sync:", path)
				go httpclient.Post(path, `{"syncStyle": "sync"}`, false)
			}

		}
	}()
	ginserver.ResponseOk(c, nil)

}

// enter
func ServiceState(c *gin.Context) {
	belogs.Info("ServiceState()")

	ssr := model.ServiceStateRequest{}
	err := c.ShouldBindJSON(&ssr)
	if err != nil {
		belogs.Error("ServiceState(): ShouldBindJSON fail :", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ServiceState(): ServiceStateRequest:", jsonutil.MarshalJson(ssr))
	serviceState, err := handleServiceState(ssr)
	if err != nil {
		belogs.Error("ServiceState(): ServiceState fail, ssr :", jsonutil.MarshalJson(ssr),
			"    serviceState:", serviceState, err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ServiceState(): serviceState:", jsonutil.MarshalJson(serviceState))
	ginserver.ResponseOk(c, *serviceState)
}

// just return valid/warning/invalid count in cer/roa/mft/crl
func Results(c *gin.Context) {
	belogs.Info("Results()")

	r, err := getResults()
	if err != nil {
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("Results():results:", jsonutil.MarshalJson(r))
	c.JSON(http.StatusOK, r)
}

func ExportRoas(c *gin.Context) {
	belogs.Info("ExportRoas()")
	r, err := exportRoas()
	if err != nil {
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ExportRoas():exportRoas:", jsonutil.MarshalJson(r))
	c.JSON(http.StatusOK, r)
}

func ExportRtrForManrs(c *gin.Context) {
	belogs.Info("ExportRtrForManrs()")
	r, err := exportRtrForManrs()
	if err != nil {
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ExportRtrForManrs():exportRtrForManrs:", jsonutil.MarshalJson(r))
	c.JSON(http.StatusOK, jsonutil.MarshallJsonIndent(r))
}

func ExportRtrForManrsConsole() (err error) {
	if len(os.Args) < 2 {
		return errors.New("no file name")
	}
	fileName := os.Args[1]

	r, err := exportRtrForManrs()
	if err != nil {
		belogs.Error("fail to get export file :", err)
		return errors.New("fail to get export file")
	}
	belogs.Debug("export number :", len(r))

	err = fileutil.WriteBytesToFile(fileName, []byte(jsonutil.MarshallJsonIndent(r)))
	if err != nil {
		belogs.Error("fail to write to export file :", err)
		return errors.New("fail to write to export file")
	}
	belogs.Info("success to export to file:", fileName)
	return nil
}
