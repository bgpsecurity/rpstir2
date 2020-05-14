package packet

import (
	"bytes"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"

	"parsevalidate/util"
)

func ExtractSkiOid(oidPackets *[]OidPacket, fileByte []byte) (ski string, err error) {
	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidSubjectKeyIdentifierKey {
			skiPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("skiPacket", skiPacket, true, false)

			belogs.Debug("ExtractSkiOid(): len(skiPPacket.Children) :", len(skiPacket.Children))
			by, _ := skiPacket.Value.([]byte)
			//logs.LogDebugBytes(("skiPacket bytes ", by)
			_, datapos, _ := util.DecodeFiniteLen(by)
			return convert.Bytes2String(by[datapos:]), nil
		}
	}
	return reExtractSkiOid(fileByte)
}

// if decode packet fail ,so try again using OID to decode
func reExtractSkiOid(fileByte []byte) (ski string, err error) {
	/*
		using OID to locate, and go to the child level
		SEQUENCE (2 elem)
		     OBJECT IDENTIFIER 2.5.29.14 subjectKeyIdentifier (X.509 extension)
		        OCTET STRING (1 elem)
		         OCTET STRING (20 byte) 943D33A631917E4FEA2F57F77D3F315744FC9127
	*/

	belogs.Debug("reExtractSkiOid():len(fileByte): ", len(fileByte))
	pos0 := bytes.Index(fileByte, oidSkiKeyByte)
	var datapos uint64 = uint64(pos0)
	var datalen uint64 = uint64(0)
	belogs.Debug("reExtractSkiOid():enum0 pos:", datapos, "   datalen:", datalen)
	if datapos <= 0 {
		return "", errors.New("not found " + oidManifestKey)
	}
	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	oct0 := fileByte[int(datapos)+len(oidSkiKeyByte):]
	//logs.LogDebugBytes(("reExtractSkiOid():oct0:", oct0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(oct0)
	belogs.Debug("reExtractSkiOid():seq0 pos:", datapos)

	//it is string, so easy , so not need use DecodePacket, just get string
	ski0 := oct0[datapos : datapos+datalen]
	//logs.LogDebugBytes(("reExtractSkiOid():ski0:", ski0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(ski0)

	skiValue := ski0[datapos : datapos+datalen]
	//logs.LogDebugBytes(("reExtractSkiOid():skiValue:", skiValue)
	return convert.Bytes2String(skiValue), nil

}
