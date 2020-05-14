package packet

import (
	"bytes"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"

	"model"
	"parsevalidate/util"
)

func ExtractMftOid(oidPackets *[]OidPacket, certFile string, fileByte []byte, mftModel *model.MftModel) (err error) {

	found := false
	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidManifestKey {

			if len(oidPacket.ParentPacket.Children) > 1 {
				seq0 := oidPacket.ParentPacket.Children[1]
				if len(seq0.Children) > 0 {
					octPacket := seq0.Children[0]
					if len(octPacket.Children) > 0 {
						secPacket := octPacket.Children[0]
						PrintPacketString("secPacket", secPacket, true, false)

						if len(secPacket.Children) > 0 {

							//var manifest = &[]Manifest{}
							var secPacketSeq *Packet
							belogs.Debug("len(secPacket.Children):", len(secPacket.Children))
							if len(secPacket.Children) > 1 {
								secPacketSeq = secPacket //.Children[0] //.Children[0]
								PrintPacketString("len(secPacket.Children) > 1 secPacketSeq", secPacketSeq, true, false)
							} else {
								secPacketSeq = secPacket.Children[0]
								PrintPacketString("len(secPacket.Children) == 0 secPacketSeq", secPacketSeq, true, false)
							}
							PrintPacketString("secPacketSeq", secPacketSeq, true, false)

							err = extractMftOidImpl(secPacketSeq, certFile, mftModel)
							found = true

						}

					}

				}

			}
		}

	}
	if found {
		return nil
	}

	//if decode packet fail ,so try again using OID to decode
	return reExtractMftOid(fileByte, certFile, mftModel)

}

// if decode packet fail ,so try again using OID to decode
func extractMftOidImpl(secPacketSeq *Packet, certFile string, mftModel *model.MftModel) (err error) {
	belogs.Debug("extractMftOidImpl():secPacketSeq:")

	manifestNumber := secPacketSeq.Children[0]
	PrintPacketString("manifestNumber", manifestNumber, true, false)
	//manifestInfo.ManifestNumber = manifestNumber.Value.(string)
	mftModel.MftNumber = convert.Bytes2String(manifestNumber.Bytes())

	thisUpdate := secPacketSeq.Children[1]
	PrintPacketString("thisUpdate", thisUpdate, true, false)
	timeStr := thisUpdate.Value.(string)
	mftModel.ThisUpdate, err = convert.String2Time(timeStr)
	belogs.Debug("extractMftOidImpl(): manifestInfo.ThisUpdate: ", mftModel.ThisUpdate, err)

	nextUpdate := secPacketSeq.Children[2]
	PrintPacketString("nextUpdate", nextUpdate, true, false)
	timeStr = nextUpdate.Value.(string)
	mftModel.NextUpdate, err = convert.String2Time(timeStr)
	belogs.Debug("extractMftOidImpl(): manifestInfo.NextUpdate: ", mftModel.NextUpdate, err)

	fileHashAlg := secPacketSeq.Children[3]
	PrintPacketString("fileHashAlg", fileHashAlg, true, false)
	mftModel.FileHashAlg = fileHashAlg.Value.(string)

	fileAndHashs := make([]model.FileHashModel, 0)

	fileList := secPacketSeq.Children[4]
	if len(fileList.Children) > 0 {
		for _, fileAndHashPacket := range fileList.Children {
			if len(fileAndHashPacket.Children) > 1 {
				fileAndHash := model.FileHashModel{}

				file := fileAndHashPacket.Children[0]
				PrintPacketString("file", file, true, false)
				fileAndHash.File, err = GetFileOrHashValue("file", file)
				belogs.Debug("extractMftOidImpl(): GetFileOrHashValue(file) fileAndHash.File: ", fileAndHash.File, err)
				if err != nil {
					continue
				}

				hash := fileAndHashPacket.Children[1]
				PrintPacketString("hash", hash, true, false)
				fileAndHash.Hash, err = GetFileOrHashValue("hash", hash)
				belogs.Debug("extractMftOidImpl(): GetFileOrHashValue(hash) fileAndHash.Hash: ", fileAndHash.Hash, err)
				if err != nil {
					continue
				}

				fileAndHashs = append(fileAndHashs, fileAndHash)

			}
		}
	}
	mftModel.FileHashModels = fileAndHashs

	return nil

}

// if decode packet fail ,so try again using OID to decode
func reExtractMftOid(fileByte []byte, certFile string, mftModel *model.MftModel) (err error) {

	/*
		 IA5String include ASCII, such as NULL,BEL,TAB,NL,LF,CR and 32~126.

				   OBJECT IDENTIFIER 1.2.840.113549.1.9.16.1.26 rpkiManifest (S/MIME Content Types)
				   [0] (1 elem)   //enum0
				     OCTET STRING (1 elem)  //oct0
				       SEQUENCE (5 elem)   //seq0
				         INTEGER 36        //int0
				         GeneralizedTime 2018-10-09 20:13:12 UTC     //gt0
				         GeneralizedTime 2018-10-10 02:13:12 UTC     //gt1
				         OBJECT IDENTIFIER 2.16.840.1.101.3.4.2.1 sha-256 (NIST Algorithm)
				         SEQUENCE (1 elem)                     //seq1
				           SEQUENCE (2 elem)                   //seq2
				             IA5String kwdO-iY3VPEY37RRTr8JDRD7J98.crl     //fil1
				             BIT STRING (256 bit) 1010000100101010010010010100011010001110111100100010101010101010110101…


				SEQUENCE (2 elem)
			        OBJECT IDENTIFIER 1.2.840.113549.1.9.16.1.26 rpkiManifest (S/MIME Content Types)
			        [0] (1 elem)
			          OCTET STRING (1 elem)
			            OCTET STRING (1 elem)
			              SEQUENCE (5 elem)
			                INTEGER 176
			                GeneralizedTime 2018-10-09 22:00:27 UTC
			                GeneralizedTime 2018-10-10 23:00:27 UTC
			                OBJECT IDENTIFIER 2.16.840.1.101.3.4.2.1 sha-256 (NIST Algorithm)
			                SEQUENCE (11 elem)
			                  SEQUENCE (2 elem)
			                    IA5String 080693385e9ce0c359cf2e5ded829dc11aa18c4

	*/
	belogs.Debug("reExtractMftOid():len(fileByte): ", len(fileByte))
	pos0 := bytes.Index(fileByte, oidManifestKeyByte)
	var datapos uint64 = uint64(pos0)
	var datalen uint64 = uint64(0)
	belogs.Debug("reExtractMftOid():enum0 pos:", datapos, "  datalen:", datalen)
	if datapos <= 0 {
		return errors.New("not found " + oidManifestKey)
	}
	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	enum0 := fileByte[int(datapos)+len(oidManifestKeyByte):]
	//logs.LogDebugBytes(("reExtractMftOid():enum0:", enum0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(enum0)
	belogs.Debug("reExtractMftOid():oct0 pos:", datapos, "  datalen:", datalen)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	oct0 := enum0[datapos:]
	//logs.LogDebugBytes(("reExtractMftOid():oct0:", oct0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(oct0)
	belogs.Debug("reExtractMftOid():seq0 pos:", datapos, "  datalen:", datalen)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	seq0 := oct0[datapos:]
	//logs.LogDebugBytes(("reExtractMftOid():seq0:", seq0)
	seq0Class := seq0[0]
	// if it is not sequence 0x30, then tray a new level
	if seq0Class != byte(0x30) {

		datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(seq0)
		belogs.Debug("reExtractMftOid():again datalen:", datalen, "    datapos:", datapos)
		seq0 = seq0[int(datapos):]
		//logs.LogDebugBytes(("reExtractMftOid():again seq0:", seq0)

	} else {

	}
	//calc length of seq, and substring
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(seq0)
	belogs.Debug("reExtractMftOid():true datalen:", datalen, "    datapos:", datapos)
	seq0 = seq0[:datapos+datalen]

	//donot care about 0x00 0x00, using datalen to get the real length
	//logs.LogDebugBytes(("reExtractMftOid():true seq0:", seq0)

	pack := DecodePacket(seq0)

	// manifests,
	return extractMftOidImpl(pack, certFile, mftModel)

}

func GetFileOrHashValue(fileOrHash string, packet *Packet) (string, error) {

	PrintPacketString(fileOrHash, packet, true, false)

	typ, _ := convert.GetInterfaceType(packet.Value)
	belogs.Debug("GetFileOrHashValue(): ", fileOrHash, typ)

	//1. try to string
	if typ == "string" {
		result := convert.ToString(packet.Value)
		belogs.Debug("GetFileOrHashValue(): ToString ", fileOrHash, result)
		if len(result) > 0 {
			return result, nil
		}
	}

	//2. is "hash" and is []byte
	belogs.Debug("GetFileOrHashValue(): fileOrHash: ", fileOrHash, "   typ:", typ)
	if fileOrHash == "hash" && typ == "[]byte" {
		if bb, p := packet.Value.([]byte); p && len(bb) > 0 {
			belogs.Debug("GetFileOrHashValue(): fileOrHash: ", fileOrHash, "   bb[0]:", bb[0])
			// is 0x00 start
			if bb[0] == 0x00 {
				bb = (bb[1:])
			}
			result := convert.Bytes2String(bb)
			belogs.Debug("GetFileOrHashValue(): Bytes2String ", fileOrHash, result)
			return result, nil

		}
	}

	//3. try byte to encode to string
	belogs.Debug("GetFileOrHashValue():cannot conver to string ", fileOrHash)
	return "", errors.New("this " + fileOrHash + " cannot conver to string")
}
