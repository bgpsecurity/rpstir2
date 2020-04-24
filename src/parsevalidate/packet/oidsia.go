package packet

import (
	"bytes"

	belogs "github.com/astaxie/beego/logs"

	"model"
	"parsevalidate/util"
)

func ExtractSiaOid(oidPackets *[]OidPacket, fileByte []byte) (subjectInfoAccess model.SiaModel, err error) {
	//oidRpkiManifestKey,oidRpkiNotifyKey,oidCaRepositoryKey,oidSignedObjectKey is children of SubjectInfoAccess
	var sia model.SiaModel = model.SiaModel{}

	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidRpkiManifestKey {
			belogs.Debug("ExtractSiaOid():found oidRpkiManifestKey:")
			oidRpkiManifestPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("oidRpkiManifestPacket", oidRpkiManifestPacket, true, false)

			bytes := oidRpkiManifestPacket.Bytes()
			_, datapos, _ := util.DecodeFiniteLen(bytes)

			//logs.LogDebugBytes(("oidRpkiManifestPacket bytes[2:]", bytes[datapos:])
			sia.RpkiManifest = string(bytes[datapos:])
			belogs.Debug("ExtractSiaOid():sia.rpkiManifest:", sia.RpkiManifest)
		}

		if oidPacket.Oid == oidRpkiNotifyKey {
			belogs.Debug("ExtractSiaOid():found oidRpkiNotifyKey:")

			oidRpkiNotifyPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("oidRpkiNotifyPacket", oidRpkiNotifyPacket, true, false)

			bytes := oidRpkiNotifyPacket.Bytes()
			_, datapos, _ := util.DecodeFiniteLen(bytes)

			//logs.LogDebugBytes(("oidRpkiNotifyPacket bytes[2:]", bytes[datapos:])
			sia.RpkiNotify = string(bytes[datapos:])
			belogs.Debug("ExtractSiaOid():sia.rpkiNotify:", sia.RpkiNotify)
		}

		if oidPacket.Oid == oidCaRepositoryKey {
			belogs.Debug("ExtractSiaOid():found oidCaRepositoryKey:")

			oidCaRepositoryPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("oidCaRepositoryPacket", oidCaRepositoryPacket, true, false)

			bytes := oidCaRepositoryPacket.Bytes()
			_, datapos, _ := util.DecodeFiniteLen(bytes)

			//logs.LogDebugBytes(("oidCaRepositoryPacket bytes[2:]", bytes[datapos:])
			sia.CaRepository = string(bytes[datapos:])
			belogs.Debug("ExtractSiaOid():sia.caRepository:", sia.CaRepository)
		}

		if oidPacket.Oid == oidSignedObjectKey {
			belogs.Debug("ExtractSiaOid():found oidSignedObjectKey:")

			oidSignedObjectPacket := oidPacket.ParentPacket.Children[1]
			PrintPacketString("oidSignedObjectPacket", oidSignedObjectPacket, true, false)

			bytes := oidSignedObjectPacket.Bytes()
			_, datapos, _ := util.DecodeFiniteLen(bytes)

			//logs.LogDebugBytes(("oidSignedObjectPacket bytes[2:]", bytes[datapos:])
			sia.SignedObject = string(bytes[datapos:])
			belogs.Debug("ExtractSiaOid():sia.signedObject:", sia.SignedObject)
		}
	}
	if len(sia.CaRepository) > 0 || len(sia.RpkiManifest) > 0 ||
		len(sia.RpkiNotify) > 0 || len(sia.SignedObject) > 0 {
		return sia, nil
	}
	return reExtractSiaOid(fileByte)
}

// if decode packet fail ,so try again using OID to decode
func reExtractSiaOid(fileByte []byte) (subjectInfoAccess model.SiaModel, err error) {
	/*
		SEQUENCE (2 elem)
		     OBJECT IDENTIFIER 1.3.6.1.5.5.7.48.11 signedObject (PKIX subject/authority info access descriptor)
		     [6] rsync://rpki.ripe.net/repository/DEFAULT/be/c37497-6376-461e-93c6-9778674edc97/1…
	*/
	var sia model.SiaModel = model.SiaModel{}
	sia.RpkiManifest, _ = reExtractSiaSubOid(oidRpkiManifestKeyByte, fileByte)
	sia.RpkiNotify, _ = reExtractSiaSubOid(oidRpkiNotifyKeyByte, fileByte)
	sia.CaRepository, _ = reExtractSiaSubOid(oidCaRepositoryKeyByte, fileByte)
	sia.SignedObject, _ = reExtractSiaSubOid(oidSignedObjectKeyByte, fileByte)
	return sia, nil

}

// if decode packet fail ,so try again using OID to decode
func reExtractSiaSubOid(oidKeyByte []byte, fileByte []byte) (sub string, err error) {
	/*
		SEQUENCE (2 elem)   //seq0
		  OBJECT IDENTIFIER 1.3.6.1.5.5.7.48.11 signedObject (PKIX subject/authority info access descriptor)   //pos0
		  [6] rsync://rpki.ripe.net/repository/DEFAULT/be/c37497-6376-461e-93c6-9778674edc97/1…   //mf0
	*/
	belogs.Debug("reExtractSiaSubOid():len(fileByte): ", len(fileByte))
	pos0 := bytes.Index(fileByte, oidKeyByte)
	belogs.Debug("reExtractSiaSubOid():enum0 pos:", pos0)
	// may not exist
	if pos0 <= 0 {
		return "", nil
	}

	var datapos uint64 = uint64(pos0)
	var datalen uint64 = uint64(0)
	belogs.Debug("reExtractSiaSubOid():datapos, datalen:", datapos, datalen)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	sub0 := fileByte[int(datapos)+len(oidKeyByte):]
	//logs.LogDebugBytes(("reExtractSiaSubOid():sub0:", sub0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(sub0)
	belogs.Debug("reExtractSiaSubOid():sub0 pos:", datapos)

	//it is string, so easy , so not need use DecodePacket, just get string
	sub0Value := sub0[datapos : datapos+datalen]
	//logs.LogDebugBytes(("reExtractSiaSubOid():sub0Value:", sub0Value)
	return string(sub0Value), nil

}
