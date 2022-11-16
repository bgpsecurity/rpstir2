package rtrclient

import (
	"bytes"
	"net"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	rtrserver "rpstir2-rtrserver"
)

type RtrTcpClientProcessFunc struct {
}

func (rq *RtrTcpClientProcessFunc) ActiveSend(conn *net.TCPConn, tcpClientProcessChan string) (err error) {
	start := time.Now()
	var rtrPduModel rtrserver.RtrPduModel
	if "resetquery" == tcpClientProcessChan {
		rtrPduModel = rtrserver.NewRtrResetQueryModel(rtrserver.PDU_PROTOCOL_VERSION_2)
	} else if "serialquery" == tcpClientProcessChan {
		rtrPduModel = rtrserver.NewRtrSerialQueryModel(rtrserver.PDU_PROTOCOL_VERSION_2, 1, 1)
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
		rtrPduModel, err := rtrserver.ParseToRtrPduModel(buf)
		if err != nil {
			return
		}
		belogs.Info("OnReceive(): client receive bytes:\n"+convert.PrintBytes(receiveData, 8)+"\n   parseTo:", jsonutil.MarshalJson(rtrPduModel), "   time(s):", time.Since(start))
	}()
	return nil

}
