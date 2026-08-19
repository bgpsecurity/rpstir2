package rtrproducer

import (
	"fmt"
	"testing"

	rtrslurm "github.com/bgpsecurity/rpstir2/rtr-producer/slurm"
	rtrsync "github.com/bgpsecurity/rpstir2/rtr-producer/sync"
	_ "github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/xormdb"
)

func TestRtrUpdateFromSync(t *testing.T) {
	// start mysql
	err := xormdb.InitMySql()
	if err != nil {

		fmt.Println("rpstir2 failed to start, ", err)
		return
	}
	defer xormdb.XormEngine.Close()
	//xormdb.XormEngine.ShowSQL(true)

	nextStep, err := rtrsync.RtrUpdateFromSync()
	fmt.Println(nextStep, err)
}

func TestRtrUpdateFromSlurm(t *testing.T) {
	// start mysql
	err := xormdb.InitMySql()
	if err != nil {

		fmt.Println("rpstir2 failed to start, ", err)
		return
	}
	defer xormdb.XormEngine.Close()
	//xormdb.XormEngine.ShowSQL(true)

	err = rtrslurm.RtrUpdateFromSlurm()
	fmt.Println(err)
}
