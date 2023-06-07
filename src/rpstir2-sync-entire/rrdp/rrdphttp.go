package rrdp

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
	model "rpstir2-model"
)

// start to rrdp from sync
func RrdpRequest(c *gin.Context) {
	belogs.Debug("RrdpRequest(): start")

	syncUrls := model.SyncUrls{}
	err := c.ShouldBindJSON(&syncUrls)
	belogs.Info("RrdpRequest(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RrdpRequest(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}

	go rrdpRequest(&syncUrls)

	ginserver.ResponseOk(c, nil)
}
