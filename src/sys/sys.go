package sys

import (
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
)

// initReset 初始化数据
//
//	@param sysStyle 初始化类型
//	@return err 返回错误
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

	err = initResetPath()
	if err != nil {
		belogs.Error("initReset(): initResetPath  fail:", err)
		return err
	}
	belogs.Debug("initReset(): initResetPath ok, reset local file cache", sysStyle)

	belogs.Info("initReset():ok", sysStyle, "  time(s):", time.Since(start))
	return nil
}

// initResetPath 初始化本地缓存目录
func initResetPath() (err error) {
	//delete repo dir
	err = os.RemoveAll(conf.String("rsync::destPath"))
	if err != nil {
		belogs.Error("initResetPath(): RemoveAll rsync::destPath fail:", conf.String("rsync::destPath"), err)
		return err
	}
	err = os.MkdirAll(conf.String("rsync::destPath"), os.ModePerm)
	if err != nil {
		belogs.Error("initResetPath(): MkdirAll rsync::destPath fail:", conf.String("rsync::destPath"), err)
		return err
	}

	//delete repo rrdpdir
	err = os.RemoveAll(conf.String("rrdp::destPath"))
	if err != nil {
		belogs.Error("initResetPath(): RemoveAll rrdp::destPath fail:", conf.String("rrdp::destPath"), err)
		return err
	}
	err = os.MkdirAll(conf.String("rrdp::destPath"), os.ModePerm)
	if err != nil {
		belogs.Error("initResetPath(): MkdirAll rrdp::destPath fail:", conf.String("rrdp::destPath"), err)
		return err
	}
	return nil
}
