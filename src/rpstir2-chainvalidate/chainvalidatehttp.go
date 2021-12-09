package chainvalidate

import (
	"github.com/cpusoft/goutil/belogs"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/gin-gonic/gin"
)

// upload file to parse
func ChainValidateStart(c *gin.Context) {

	rpstir2Url := "https://" + conf.String("rpstir2-rp::serverHost") + ":" + conf.String("rpstir2-rp::serverHttpsPort")
	rpstir2VcUrl := "https://" + conf.String("rpstir2-vc::serverHost") + ":" + conf.String("rpstir2-vc::serverHttpsPort")
	belogs.Info("ChainValidateStart(): start,  rpstir2Url:", rpstir2Url, "   rpstir2VcUrl:", rpstir2VcUrl)

	//check serviceState
	httpclient.Post(rpstir2Url+"/sys/servicestate", `{"operate":"enter","state":"chainvalidate"}`, false)

	go func() {
		nextStep, err := chainValidateStart()
		belogs.Debug("ChainValidateStart():  ChainValidateStart end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("ParseValidateStart():  chainValidateStart fail", err)
			httpclient.Post(rpstir2Url+"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {
			// leave serviceState
			httpclient.Post(rpstir2Url+"/sys/servicestate", `{"operate":"leave","state":"chainvalidate"}`, false)

			// rtr producer
			go httpclient.Post(rpstir2VcUrl+"/rtrproducer/updatefromsync", `{"lastStep":"chainValidateStart"}`, false)

			// call statistics
			go httpclient.Post(rpstir2Url+"/statistic/start", "", false)

			// call roacompete
			go httpclient.Post(rpstir2Url+"/roacompete/start", "", false)

			// call roahistory
			//go httpclient.Post(rpstir2Url+"/roahistory/start", "", false)
			belogs.Info("ParseValidateStart(): end,  nextStep is :", nextStep)
		}
	}()

	ginserver.ResponseOk(c, nil)

}
