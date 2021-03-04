package sys

import (
	"os"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	db "sys/db"
	sysmodel "sys/model"
)

//
func InitReset(sysStyle sysmodel.SysStyle) (err error) {
	start := time.Now()
	belogs.Debug("InitReset():will InitReset db, sysStyle:", jsonutil.MarshalJson(sysStyle))

	// reset db
	err = db.InitResetDb(sysStyle)
	if err != nil {
		belogs.Error("InitReset():InitReset db fail:", err)
		return err
	}
	belogs.Debug("InitReset(): InitReset db ok, will reset local file cache", sysStyle)

	//delete repo dir
	os.RemoveAll(conf.VariableString("rsync::destPath"))
	os.MkdirAll(conf.VariableString("rsync::destPath"), os.ModePerm)

	//delete repo rrdpdir
	os.RemoveAll(conf.VariableString("rrdp::destPath"))
	os.MkdirAll(conf.VariableString("rrdp::destPath"), os.ModePerm)

	if sysStyle.SysStyle == "fullsync" {
		go func() {
			// default syncStyle is sync
			// but ,if it get syncStyle from sysStyle, it will return to "/sync/start"
			syncStyle := "sync"
			if len(sysStyle.SyncStyle) > 0 {
				syncStyle = sysStyle.SyncStyle
			}
			belogs.Info("InitReset():fullsync will call sync:", syncStyle)
			httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
				"/sync/start", `{"syncStyle": "`+syncStyle+`"}`, false)
		}()
	}
	belogs.Info("InitReset():ok", sysStyle, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
