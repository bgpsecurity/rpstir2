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
	belogs.Debug("initReset():will InitReset, sysStyle:", jsonutil.MarshalJson(sysStyle))

	// reset db
	err = initResetDb(sysStyle)
	if err != nil {
		belogs.Error("initReset(): initResetDb  fail:", err)
		return err
	}
	belogs.Debug("initReset(): initResetDb ok, will reset local file cache", sysStyle)

	initResetPath()
	belogs.Debug("initReset(): initResetPath ok, reset local file cache", sysStyle)

	belogs.Info("initReset():ok", sysStyle, "  time(s):", time.Since(start))
	return nil
}

func initResetPath() {
	//delete repo dir
	os.RemoveAll(conf.VariableString("rsync::destPath"))
	os.MkdirAll(conf.VariableString("rsync::destPath"), os.ModePerm)

	//delete repo rrdpdir
	os.RemoveAll(conf.VariableString("rrdp::destPath"))
	os.MkdirAll(conf.VariableString("rrdp::destPath"), os.ModePerm)
}
