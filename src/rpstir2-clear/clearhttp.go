package clear

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/gin-gonic/gin"
)

func ClearStart(c *gin.Context) {
	belogs.Info("ClearStart(): start: ")

	go clearStart()

	ginserver.ResponseOk(c, nil)
}
