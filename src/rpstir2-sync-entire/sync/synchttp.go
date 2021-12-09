package sync

import (
	"errors"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/gin-gonic/gin"
)

// start to sync
func SyncStart(c *gin.Context) {
	belogs.Info("SyncStart(): start")

	// get syncStyle
	syncStyle := model.SyncStyle{}
	err := c.ShouldBindJSON(&syncStyle)
	if err != nil {
		belogs.Error("SyncStart(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	if syncStyle.SyncStyle != "sync" && syncStyle.SyncStyle != "rrdp" && syncStyle.SyncStyle != "rsync" {
		belogs.Error("SyncStart(): syncStyle should be sync or rrdp or rsyncc, it is ", syncStyle.SyncStyle)
		ginserver.ResponseFail(c, errors.New("SyncStyle should be sync or rrdp or rsync"), "")
		return
	}
	belogs.Debug("SyncStart(): syncStyle:", syncStyle)

	//check serviceState
	ssr := model.ServiceState{}
	err = httpclient.PostAndUnmarshalResponseModel("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"sync"}`, false, &ssr)
	if err != nil {
		belogs.Error("SyncStart(): PostAndUnmarshalResponseModel failed, err:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}

	go func() {
		nextStep, err := syncStart(syncStyle)
		belogs.Debug("SyncStart(): syncStart end,  nextStep is :", nextStep, err)

		if err != nil {
			// will end this whole sync
			belogs.Error("SyncStart(): SyncStart fail,  syncStyle is :", syncStyle, err)
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {

			// will go next step
			if nextStep == "fullsync" {
				// leave current sync now, and start new fullsync
				httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
					"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
				//{"sysStyle": "fullsync","syncPolicy":"entire"}
				go httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
					"/sys/initreset", `{"sysStyle":"fullsync", "syncPolicy":"entire", "syncStyle":"`+syncStyle.SyncStyle+`"}`, false)
			} else if nextStep == "parsevalidate" {
				// will end sync ,and will start next step
				httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
					"/sys/servicestate", `{"operate":"leave","state":"sync"}`, false)

				go httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
					"/parsevalidate/start", "", false)
			}
			belogs.Info("SyncStart(): end, nextStep is :", nextStep)
		}

	}()
	ginserver.ResponseOk(c, nil)
}

// get result from rsync
func RsyncResult(c *gin.Context) {
	belogs.Debug("RsyncResult(): start")

	r := model.SyncResult{}
	err := c.ShouldBindJSON(&r)
	if err != nil {
		belogs.Error("RsyncResult(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	rsyncResult(&r)
	ginserver.ResponseOk(c, nil)
}

// get result from rrdp
func RrdpResult(c *gin.Context) {
	belogs.Debug("RrdpResult(): start")

	r := model.SyncResult{}
	err := c.ShouldBindJSON(&r)
	if err != nil {
		belogs.Error("RrdpResult(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	rrdpResult(&r)
	ginserver.ResponseOk(c, nil)
}

// sync from local, for history repo data: need reset and just start from diff, then parse....
func SyncLocalStart(c *gin.Context) {
	belogs.Info("SyncLocalStart(): start")
	go LocalStart()
	ginserver.ResponseOk(c, nil)
}
