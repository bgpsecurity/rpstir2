package tal

import (
	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

//
func GetTals(c *gin.Context) {
	belogs.Info("GetTals")

	talModels, err := getTals()
	belogs.Debug("GetTals(): getTals, talModels:", jsonutil.MarshalJson(talModels))
	if err != nil {
		ginserver.ResponseFail(c, err, "")
	} else {
		talModelsResponse := model.TalModelsResponse{TalModels: talModels}
		ginserver.ResponseOk(c, talModelsResponse)
	}
}
