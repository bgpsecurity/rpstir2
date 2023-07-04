package rsync

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
	model "rpstir2-model"
)

// start to rsync from sync
func RsyncRequest(c *gin.Context) {
	belogs.Debug("RsyncRequest(): start")

	syncUrls := model.SyncUrls{}
	err := c.ShouldBindJSON(&syncUrls)
	belogs.Info("RsyncRequest(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RsyncRequest(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}

	go rsyncRequest(&syncUrls)

	ginserver.ResponseOk(c, nil)
}

// start local rsync
func LocalRsyncRequest(c *gin.Context) {
	belogs.Debug("LocalRsyncRequest(): start")

	syncUrls := model.SyncUrls{}
	err := c.ShouldBindJSON(&syncUrls)
	belogs.Info("LocalRsyncRequest(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("LocalRsyncRequest(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}

	rsyncResult, err := localRsyncRequest(&syncUrls)
	if err != nil {
		belogs.Error("LocalRsyncRequest(): localRsyncRequest:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}
	belogs.Debug("LocalRsyncRequest(): rsyncResult:", rsyncResult)

	ginserver.ResponseOk(c, rsyncResult)

}
