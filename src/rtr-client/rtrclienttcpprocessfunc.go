package rtrclient

import (
	"bytes"
	"net"
	"time"

	rtrcore "github.com/bgpsecurity/rpstir2/rtr-core"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrTcpClientProcessFunc struct {
}

func (rq *RtrTcpClientProcessFunc) ActiveSend(conn *net.TCPConn, tcpClientProcessChan string) (err error) {
	start := time.Now()
	rtrProtocolVersion := uint8(conf.Int("rtr::protocolVersion"))
	belogs.Debug("RtrTcpClientProcessFunc.ActiveSend(): rtrProtocolVersion:", rtrProtocolVersion)
	var rtrPduModel rtrcore.RtrPduModel
	if "resetquery" == tcpClientProcessChan {
		rtrPduModel = rtrcore.NewRtrResetQueryModel(rtrProtocolVersion)
	} else if "serialquery" == tcpClientProcessChan {
		rtrPduModel = rtrcore.NewRtrSerialQueryModel(rtrProtocolVersion, rtrClientSerialQueryModel.SessionId, rtrClientSerialQueryModel.SerialNumber)
	}
	sendBytes := rtrPduModel.Bytes()
	belogs.Debug("ActiveSend():client:", convert.Bytes2String(sendBytes))

	_, err = conn.Write(sendBytes)
	if err != nil {
		belogs.Debug("ActiveSend():client:  conn.Write() fail,  ", convert.Bytes2String(sendBytes), err)
		return err
	}
	belogs.Info("ActiveSend(): client send:", jsonutil.MarshalJson(rtrPduModel), "   time(s):", time.Since(start))
	return nil

}
func (rq *RtrTcpClientProcessFunc) OnReceive(conn *net.TCPConn, receiveData []byte) (err error) {

	go func() {
		start := time.Now()
		belogs.Debug("OnReceive():client,bytes\n" + convert.PrintBytes(receiveData, 8))
		buf := bytes.NewReader(receiveData)
		rtrPduModel, err := rtrcore.ParseToRtrPduModel(buf)
		if err != nil {
			// return
			belogs.Error("OnReceive(): client, ParseToRtrPduModel fail, will next")
		}
		belogs.Info("OnReceive(): client receive bytes:\n"+convert.PrintBytes(receiveData, 8)+"\n   parseTo:", jsonutil.MarshalJson(rtrPduModel), "   time(s):", time.Since(start))
	}()
	return nil

}
