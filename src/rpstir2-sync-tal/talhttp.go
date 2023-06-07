package tal

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
	model "rpstir2-model"
)

//
func GetTals(c *gin.Context) {
	belogs.Info("GetTals")

	talModels, err := getTals()
	if err != nil {
		belogs.Error("GetTals(): getTals fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("GetTals(): getTals, talModels:", jsonutil.MarshalJson(talModels))
	talModelsResponse := model.TalModelsResponse{TalModels: talModels}
	ginserver.ResponseOk(c, talModelsResponse)

}
