package rtrproducer

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/gin-gonic/gin"
	rtrslurm "rpstir2-rtrproducer/slurm"
	rtrsync "rpstir2-rtrproducer/sync"
)

// start to update
func RtrUpdateFromSync(c *gin.Context) {
	belogs.Info("RtrUpdateFromSync(): http start")

	//check serviceState
	httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"rtr"}`, false)

	go func() {
		nextStep, err := rtrsync.RtrUpdateFromSync()
		belogs.Debug("RtrUpdateFromSync(): http RtrUpdateFromSync end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("RtrUpdateFromSync():http RtrUpdateFromSync fail", err)
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
			belogs.Debug("RtrUpdateFromSync(): http nextStep:", nextStep, "  path:", path)
			// call transfer to push incremental
			go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::transferHttpsPort")+
				path, `{"lastStep":"rtrUpdateFromSync"}`, false)

			belogs.Info("RtrUpdateFromSync(): http RtrUpdateFromSync end,  nextStep is :", nextStep)
		}

	}()

	ginserver.ResponseOk(c, nil)
}

// start to update
func RtrUpdateFromSlurm(c *gin.Context) {
	belogs.Debug("RtrUpdateFromSlurm(): http start")

	err := rtrslurm.RtrUpdateFromSlurm()
	// leave serviceState
	if err != nil {
		// will end this whole sync
		belogs.Error("RtrUpdateFromSlurm(): http  rtrUpdateFromSlurm fail", err)
		ginserver.ResponseFail(c, err, "")
		return
	} else {

		belogs.Info("RtrUpdateFromSlurm(): http ok: will call /rtr/server/sendserialnotify, ",
			" and call /rushtransfer/triggerpushincr")
		// call serial notify to rtr client
		go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::serverHttpsPort")+
			"/rtr/server/sendserialnotify", "", false)

		// call transfer to push incremental
		go httpclient.Post("https://"+conf.String("rpstir2-vc::serverHost")+":"+conf.String("rpstir2-vc::transferHttpsPort")+
			"/rushtransfer/triggerpushincr", `{"lastStep":"rtrUpdateFromSlurm"}`, false)

		belogs.Info("RtrUpdateFromSlurm(): http  rtrUpdateFromSlurm end:")
		ginserver.ResponseOk(c, nil)
		return
	}

}
