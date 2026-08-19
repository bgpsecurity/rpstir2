package rtrserver

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
)

// server send notify to client
func ServerSendSerialNotify(c *gin.Context) {
	belogs.Info("ServerSendSerialNotify(): start")

	err := SendSerialNotify()
	if err != nil {
		belogs.Error("ServerSendSerialNotify(): SendSerialNotify: err:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}
