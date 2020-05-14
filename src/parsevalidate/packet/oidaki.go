package packet

import (
	"bytes"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"

	"parsevalidate/util"
)

func ExtractAkiOid(oidPackets *[]OidPacket, fileByte []byte) (aki string, err error) {
	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidAuthorityKeyIdentifierKey {
			akiPPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("akiPPacket", akiPPacket, true, false)

			akiPacket := akiPPacket.Children[0].Children[0]
			PrintPacketString("akiPacket", akiPacket, true, false)

			bytes := akiPacket.Bytes()
			lens := int(bytes[1])
			belogs.Debug("ExtractAkiOid akiPacket.ClassType:", akiPacket.ClassType, "   lens:", lens)
			if akiPacket.ClassType == ClassContext {
				_, datapos, _ := util.DecodeFiniteLen(bytes)
				//logs.LogDebugBytes(("akiPacket bytes[datapos:]", bytes[datapos:])
				return convert.Bytes2String(bytes[datapos:]), nil
			}
		}
	}
	return reExtractAkiOid(fileByte)
}

// if decode packet fail ,so try again using OID to decode
func reExtractAkiOid(fileByte []byte) (aki string, err error) {
	/*
		using OID to locate, and go to the child level
		   SEQUENCE (2 elem)
		     OBJECT IDENTIFIER 2.5.29.35 authorityKeyIdentifier (X.509 extension)
		     OCTET STRING (1 elem)  //oct0
		       SEQUENCE (1 elem)    //seq0
		         [0] (20 byte) 93074EFA263754F118DFB4514EBF090D10FB27DF
	*/

	belogs.Debug("reExtractAkiOid():len(fileByte): ", len(fileByte))
	pos0 := bytes.Index(fileByte, oidAkiKeyByte)
	var datapos uint64 = uint64(pos0)
	var datalen uint64 = uint64(0)
	belogs.Debug("reExtractAkiOid():enum0 pos:", datapos, "  datalen:", datalen)
	if datapos <= 0 {
		return "", errors.New("not found " + oidManifestKey)
	}
	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	oct0 := fileByte[int(datapos)+len(oidAkiKeyByte):]
	//logs.LogDebugBytes(("reExtractAkiOid():oct0:", oct0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(oct0)
	belogs.Debug("reExtractMftOid():seq0 datalen:", datalen, " datapos:", datapos)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	seq0 := oct0[datapos:]
	//logs.LogDebugBytes(("reExtractAkiOid():seq0:", seq0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(seq0)
	belogs.Debug("reExtractAkiOid():seq0 datalen:", datalen, " datapos:", datapos)

	//it is string, so easy , so not need use DecodePacket, just get string
	aki0 := seq0[datapos : datapos+datalen]
	//logs.LogDebugBytes(("reExtractAkiOid():aki0:", aki0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(aki0)

	akiValue := aki0[datapos : datapos+datalen]
	//logs.LogDebugBytes(("reExtractAkiOid():akiValue:", akiValue)
	return convert.Bytes2String(akiValue), nil

}
