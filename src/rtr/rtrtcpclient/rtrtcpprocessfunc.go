package rtrtcpclient

import (
	"bytes"
	"net"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
	rtrtcp "rtr/rtrtcp"
)

type RtrTcpClientProcessFunc struct {
}

func (rq *RtrTcpClientProcessFunc) ActiveSend(conn *net.TCPConn, tcpClientProcessChan string) (err error) {
	start := time.Now()
	var rtrPduModel rtrmodel.RtrPduModel
	if "resetquery" == tcpClientProcessChan {
		rtrPduModel = rtrmodel.NewRtrResetQueryModel(1)
	} else if "serialquery" == tcpClientProcessChan {
		rtrPduModel = rtrmodel.NewRtrSerialQueryModel(1, 1, 1)
	}
	sendBytes := rtrPduModel.Bytes()
	belogs.Debug("ActiveSend():client:", convert.Bytes2String(sendBytes))

	_, err = conn.Write(sendBytes)
	if err != nil {
		belogs.Debug("ActiveSend():client:  conn.Write() fail,  ", convert.Bytes2String(sendBytes), err)
		return err
	}
	belogs.Info("ActiveSend(): client send:", jsonutil.MarshalJson(rtrPduModel), "   time(s):", time.Now().Sub(start).Seconds())
	return nil

}
func (rq *RtrTcpClientProcessFunc) OnReceive(conn *net.TCPConn, receiveData []byte) (err error) {

	go func() {
		start := time.Now()
		belogs.Debug("OnReceive():client,", convert.Bytes2String(receiveData))
		buf := bytes.NewReader(receiveData)
		rtrPduModel, err := rtrtcp.ParseToRtrPduModel(buf)
		if err != nil {
			return
		}
		belogs.Info("OnReceive(): client receive :", jsonutil.MarshalJson(rtrPduModel), "   time(s):", time.Now().Sub(start).Seconds())
	}()
	return nil

}
