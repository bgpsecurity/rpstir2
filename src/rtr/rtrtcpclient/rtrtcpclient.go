package main

import (
	"net"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	convert "github.com/cpusoft/goutil/convert"
	_ "github.com/cpusoft/goutil/logs"
	"github.com/cpusoft/goutil/tcpudputil"

	rtrmodel "rtr/model"
)

func main() {
	belogs.Info("startTcpServer(): start")
	tcpudputil.CreateTcpClient(conf.String("rpstir2::rtrtcpserver")+":"+conf.String("rtr::tcpport"), SerialQueryDiffSessionIdProcess)
	select {}
}

func ResetQueryProcess(conn net.Conn) error {
	defer conn.Close()

	reset := rtrmodel.NewRtrResetQueryModel(1)
	belogs.Debug("ResetQueryProcess():", reset.PrintBytes())

	conn.Write(reset.Bytes())

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			belogs.Error("ResetQueryProcess(): err:", err)
			return err
		}
		belogs.Debug("ResetQueryProcess(): Read n:", n)
		recvByte := buffer[0:n]
		belogs.Debug("ResetQueryProcess(): recvByte:\r\n", convert.PrintBytes(recvByte, 8))
	}
	return nil
}

func SerialQueryProcess(conn net.Conn) error {
	defer conn.Close()

	reset := rtrmodel.NewRtrSerialQueryModel(1, 1, 1)
	belogs.Debug("SerialQueryProcess():", reset.PrintBytes())

	conn.Write(reset.Bytes())

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			belogs.Error("SerialQueryProcess(): err:", err)
			return err
		}
		belogs.Debug("SerialQueryProcess(): Read n:", n)
		recvByte := buffer[0:n]
		belogs.Debug("SerialQueryProcess(): recvByte:\r\n", convert.PrintBytes(recvByte, 8))
	}
	return nil
}

func SerialQueryDiffSessionIdProcess(conn net.Conn) error {
	defer conn.Close()

	reset := rtrmodel.NewRtrSerialQueryModel(1, 3, 1)
	belogs.Debug("SerialQueryProcess():", reset.PrintBytes())

	conn.Write(reset.Bytes())

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			belogs.Error("SerialQueryProcess(): err:", err)
			return err
		}
		belogs.Debug("SerialQueryProcess(): Read n:", n)
		recvByte := buffer[0:n]
		belogs.Debug("SerialQueryProcess(): recvByte:\r\n", convert.PrintBytes(recvByte, 8))
	}
	return nil
}
