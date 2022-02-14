package rsync

import (
	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
)

// start to rsync from sync
func RsyncStart(c *gin.Context) {
	belogs.Debug("RsyncStart(): start")

	syncUrls := model.SyncUrls{}
	err := c.ShouldBindJSON(&syncUrls)
	belogs.Debug("RsyncStart(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RsyncStart(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}

	go rsyncStart(&syncUrls)

	ginserver.ResponseOk(c, nil)
}

// start local rsync
func RsyncLocalStart(c *gin.Context) {
	belogs.Debug("RsyncLocalStart(): start")

	syncUrls := model.SyncUrls{}
	err := c.ShouldBindJSON(&syncUrls)
	belogs.Debug("RsyncStart(): syncUrls:", syncUrls, err)
	if err != nil {
		belogs.Error("RsyncStart(): ShouldBindJSON:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}

	rsyncResult, err := LocalStart(&syncUrls)
	if err != nil {
		belogs.Error("RsyncLocalStart(): LocalStart:", err)
		ginserver.ResponseFail(c, err, nil)
		return
	}
	belogs.Debug("RsyncLocalStart(): rsyncResult:", rsyncResult)

	ginserver.ResponseOk(c, rsyncResult)

}
