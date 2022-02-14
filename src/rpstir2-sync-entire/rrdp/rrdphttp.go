package rrdp

import (
	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
)

// start to rrdp from sync
func RrdpStart(c *gin.Context) {
	belogs.Debug("RrdpStart(): start")

	syncUrls := model.SyncUrls{}
	err := c.ShouldBindJSON(&syncUrls)
	belogs.Debug("RrdpStart(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RrdpStart(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}

	go rrdpStart(&syncUrls)

	ginserver.ResponseOk(c, nil)
}
