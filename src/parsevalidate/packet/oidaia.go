package packet

import (
	"bytes"
	"errors"

	belogs "github.com/astaxie/beego/logs"

	"model"
	"parsevalidate/util"
)

func ExtractAiaOid(oidPackets *[]OidPacket, fileByte []byte) (authorityInfoAccess model.AiaModel, err error) {
	//oidCaIssuersKey is child of AuthorityInfoAccess
	var aia model.AiaModel = model.AiaModel{}

	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidCaIssuersKey {
			caIssuersPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("akiPPacket", caIssuersPacket, true, false)

			bytes := caIssuersPacket.Bytes()
			_, datapos, _ := util.DecodeFiniteLen(bytes)

			//logs.LogDebugBytes(("akiPacket bytes[datapos:]", bytes[datapos:])
			aia.CaIssuers = string(bytes[datapos:])

			return aia, nil
		}
	}
	return reExtractAiaOid(fileByte)
}

// if parse failed by Packet, then use OID to parse again
func reExtractAiaOid(fileByte []byte) (authorityInfoAccess model.AiaModel, err error) {
	/*
		use OID, the parse next level
		   SEQUENCE (2 elem)
		     OBJECT IDENTIFIER 1.3.6.1.5.5.7.48.2 caIssuers (PKIX subject/authority info access descriptor)
		     [6] rsync://repository.lacnic.net/rpki/lacnic/48f083bb-f603-4893-9990-0284c04ceb85/8…
	*/
	var aia model.AiaModel = model.AiaModel{}
	belogs.Debug("reExtractAiaOid():len(fileByte): ", len(fileByte))
	pos0 := bytes.Index(fileByte, oidCaIssuersKeyByte)
	var datapos uint64 = uint64(pos0)
	var datalen uint64 = uint64(0)
	belogs.Debug("reExtractAiaOid():enum0 pos:", datapos, "   datalen:", datalen)
	if datapos <= 0 {
		return aia, errors.New("not found " + oidManifestKey)
	}
	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	aia0 := fileByte[int(datapos)+len(oidCaIssuersKeyByte):]
	//logs.LogDebugBytes(("reExtractAiaOid():aia0:", aia0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(aia0)
	belogs.Debug("reExtractAiaOid():aia0 pos:", datapos)

	//it is string, so easy , so not need use DecodePacket, just get string
	aiaValue := aia0[datapos : datapos+datalen]
	//logs.LogDebugBytes(("reExtractAiaOid():aiaValue:", aiaValue)
	aia.CaIssuers = string(aiaValue)
	return aia, nil

}
