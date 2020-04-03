package sys

import (
	"os"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"

	db "sys/db"
)

func InitReset(isInit bool) {
	belogs.Info("InitReset():", isInit)

	// reset db
	err := db.InitResetDb(isInit)
	if err != nil {
		belogs.Error("InitReset():InitResetDb fail:", err)
		return
	}
	belogs.Info("InitReset(): InitResetDb ok")

	//delete repo dir
	os.RemoveAll(conf.VariableString("rsync::destpath"))
	os.MkdirAll(conf.VariableString("rsync::destpath"), os.ModePerm)

	//delete repo rrdpdir
	os.RemoveAll(conf.VariableString("rrdp::destpath"))
	os.MkdirAll(conf.VariableString("rrdp::destpath"), os.ModePerm)

	belogs.Info("InitReset():ok")
}
