package sync

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func TestMain(m *testing.M) {

	initLogger()
	initMysql()
	m.Run()
}

func initLogger() {

	logConfig := make(map[string]interface{})
	logConfig["daily"] = true
	logConfig["hourly"] = false
	logConfig["maxlines"] = 0
	logConfig["maxfiles"] = 0
	logConfig["maxsize"] = 0
	logConfig["maxdays"] = 30
	logConfig["maxhours"] = 0
	logConfig["level"] = belogs.LevelDebug

	logConfigStr, _ := json.Marshal(logConfig)
	err := belogs.SetLogger(belogs.AdapterConsole, string(logConfigStr))
	if err != nil {
		fmt.Println(" SetLogger failed, " + err.Error() + ",   " + string(logConfigStr))
	}
	belogs.EnableFuncCallDepth(true)
	belogs.Debug("initLogger success")
}

func initMysql() {

	// start mysql
	err := xormdb.InitMySql()
	if err != nil {
		belogs.Error("startWebServer(): start InitMySql failed:", err)
		fmt.Println("rpstir2 failed to start, ", err)
		return
	}
	//defer xormdb.XormEngine.Close()

	belogs.Debug("initMysql success")
}
