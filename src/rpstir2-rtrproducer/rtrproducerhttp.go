package rtrproducer

import (
	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

// start to update
func RtrUpdateFromSync(c *gin.Context) {
	belogs.Info("RtrUpdateFromSync(): start")

	//check serviceState
	httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"rtr"}`, false)

	go func() {
		nextStep, err := rtrUpdateFromSync()
		belogs.Debug("RtrUpdateFromSync():  rtrUpdateFromSync end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("RtrUpdateFromSync():rtrUpdateFromSync fail", err)
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {
			// leave serviceState
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"rtr"}`, false)

			go httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/clear/start", ``, false)

			// call serial notify to rtr client
			go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::serverHttpsPort")+
				"/rtr/server/sendserialnotify", "", false)

			path := ""
			if nextStep == "full" {
				path = "/rushtransfer/triggerpushfull"
			} else if nextStep == "incr" {
				path = "/rushtransfer/triggerpushincr"
			}
			belogs.Debug("RtrUpdateFromSync():  nextStep:", nextStep, "  path:", path)
			// call transfer to push incremental
			go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::transferHttpsPort")+
				path, `{"lastStep":"rtrUpdateFromSync"}`, false)

			belogs.Info("RtrUpdateFromSync():  rtrUpdateFromSync end,  nextStep is :", nextStep)
		}

	}()

	ginserver.ResponseOk(c, nil)
}

// start to update
func RtrUpdateFromSlurm(c *gin.Context) {
	belogs.Debug("RtrUpdateFromSlurm(): start")

	scheduleModel := model.ScheduleModel{}
	c.ShouldBindJSON(&scheduleModel)
	belogs.Info("RtrUpdateFromSlurm(): scheduleModel:", jsonutil.MarshalJson(scheduleModel))

	err := rtrUpdateFromSlurm()
	// leave serviceState
	if err != nil {
		// will end this whole sync
		belogs.Error("RtrUpdateFromSlurm():  rtrUpdateFromSlurm fail", err)
		ginserver.ResponseFail(c, err, "")
		return
	} else {

		belogs.Info("RtrUpdateFromSlurm(): ok: will call /rtr/server/sendserialnotify, ",
			" and may be call /rushtransfer/triggerpushincr,  scheduleModel:", jsonutil.MarshalJson(scheduleModel))
		// call serial notify to rtr client
		go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::serverHttpsPort")+
			"/rtr/server/sendserialnotify", "", false)

		if !(scheduleModel.LastStep == "receiveIncr" ||
			scheduleModel.LastStep == "receiveFull") {
			// call transfer to push incremental
			go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::transferHttpsPort")+
				"/rushtransfer/triggerpushincr", `{"lastStep":"rtrUpdateFromSlurm"}`, false)
		}
		belogs.Info("RtrUpdateFromSlurm():  rtrUpdateFromSlurm end:")
		ginserver.ResponseOk(c, nil)
		return
	}

}
