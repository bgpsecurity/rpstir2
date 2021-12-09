package sys

import (
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
)

//
func initReset(sysStyle SysStyle) (err error) {
	start := time.Now()
	belogs.Debug("initReset():will InitReset db, sysStyle:", jsonutil.MarshalJson(sysStyle))

	// reset db
	err = InitResetDb(sysStyle)
	if err != nil {
		belogs.Error("initReset():InitReset db fail:", err)
		return err
	}
	belogs.Debug("initReset(): InitReset db ok, will reset local file cache", sysStyle)

	//delete repo dir
	os.RemoveAll(conf.VariableString("rsync::destPath"))
	os.MkdirAll(conf.VariableString("rsync::destPath"), os.ModePerm)

	//delete repo rrdpdir
	os.RemoveAll(conf.VariableString("rrdp::destPath"))
	os.MkdirAll(conf.VariableString("rrdp::destPath"), os.ModePerm)

	belogs.Info("initReset():ok", sysStyle, "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
