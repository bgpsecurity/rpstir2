package rtrclient

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
)

/*
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
*/

// start client
func ClientStart(c *gin.Context) {
	belogs.Info("ClientStart(): start")
	rtrClientStartModel := RtrClientStartModel{}
	c.ShouldBindJSON(&rtrClientStartModel)
	belogs.Debug("ClientStart(): rtrClientStartModel:", rtrClientStartModel)

	go clientStart(rtrClientStartModel)
	ginserver.ResponseOk(c, nil)
}

// stop client
func ClientStop(c *gin.Context) {
	belogs.Info("ClientEnd(): start")

	go clientStop()
	ginserver.ResponseOk(c, nil)
}

// client send serial query to server
func ClientSendSerialQuery(c *gin.Context) {
	belogs.Info("ClientSendSerialQuery(): start")

	err := clientSendSerialQuery()
	if err != nil {
		belogs.Error("ClientSendSerialQuery(): clientSendSerialQuery: err:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}

// client send reset query to server
func ClientSendResetQuery(c *gin.Context) {
	belogs.Info("ClientSendResetQuery(): start")

	err := clientSendResetQuery()
	if err != nil {
		belogs.Error("ClientSendResetQuery(): clientSendResetQuery: err:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	ginserver.ResponseOk(c, nil)
}
