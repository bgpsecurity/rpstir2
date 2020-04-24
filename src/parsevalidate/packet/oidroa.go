package packet

import (
	"bytes"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	iputil "github.com/cpusoft/goutil/iputil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	"parsevalidate/util"
)

func ExtractRoaOid(oidPackets *[]OidPacket, certFile string, fileByte []byte, roaModel *model.RoaModel) (err error) {
	belogs.Debug("ExtractRoaOid(): oidPackets.len :%d", len(*oidPackets))
	found := false
	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidRoaKey {

			if len(oidPacket.ParentPacket.Children) > 1 {
				seq0 := oidPacket.ParentPacket.Children[1]
				PrintPacketString("ExtractRoaOid():  seq0", seq0, true, false)

				if len(seq0.Children) > 0 {
					for _, octPacket := range seq0.Children {
						PrintPacketString("ExtractRoaOid():  octPacket", octPacket, true, false)
						belogs.Debug("ExtractRoaOid():len(octPacket.Children):", len(octPacket.Children))

						if len(octPacket.Children) > 0 {
							for _, routeOriginAttestationPack := range octPacket.Children {
								PrintPacketString("routeOriginAttestationPack", routeOriginAttestationPack, true, false)
								belogs.Debug("ExtractRoaOid():len(routeOriginAttestationPack.Children):", len(routeOriginAttestationPack.Children))
								// when it is indefinite length，roa parse is always wrong, may be include more data
								// when it is normal ,it will 2,or 1, or 0, cannot be 3
								if len(routeOriginAttestationPack.Children) >= 0 && len(routeOriginAttestationPack.Children) < 3 {
									if len(routeOriginAttestationPack.Children) == 1 {
										routeOriginAttestationPack = routeOriginAttestationPack.Children[0]
									} else if len(routeOriginAttestationPack.Children) == 0 {
										routeOriginAttestationPack = routeOriginAttestationPack.Parent
									}
									//if it is not integer, it will error to calc length, just give up
									if len(routeOriginAttestationPack.Children) == 0 ||
										routeOriginAttestationPack.Children[0].Tag != TagInteger {
										belogs.Debug("ExtractRoaOid():len(routeOriginAttestationPack.Children)==0 || asId.Tag != TagInteger ")
										continue
									}

									err = extractRoaOidImpl(routeOriginAttestationPack, certFile, roaModel)
									found = true

								}

							}

						}
					}

				}

			}
		}

	}
	if found {
		return nil
	}
	return reExtractRoaOid(fileByte, certFile, roaModel)
}

// if decode packet fail ,so try again using OID to decode
func extractRoaOidImpl(secPacketSeq *Packet, certFile string, roaModel *model.RoaModel) (err error) {
	belogs.Debug("extractRoaOidImpl():secPacketSeq:")

	asId := secPacketSeq.Children[0]
	PrintPacketString("asId", asId, true, false)

	ipAddrBlocks := secPacketSeq.Children[1]
	PrintPacketString("ipAddrBlocks", ipAddrBlocks, true, false)

	roaModel.Asn = asId.Value.(int64)

	if len(ipAddrBlocks.Children) > 0 {
		var ipType int

		roaIpAddressModels := make([]model.RoaIpAddressModel, 0)
		for _, roaIPAddressFamilyPack := range ipAddrBlocks.Children {
			PrintPacketString("roaIPAddressFamilyPack", roaIPAddressFamilyPack, true, false)
			if len(roaIPAddressFamilyPack.Children) > 1 {

				ipTypeByte, err := convert.Interface2Bytes(roaIPAddressFamilyPack.Children[0].Value)
				if err != nil {
					belogs.Error(err.Error() + ":" + certFile + " to hash,  bytes:" + convert.Bytes2String(roaIPAddressFamilyPack.Children[0].Bytes()))
					continue
				}
				//ipTypeByte := roaIPAddressFamilyPack.Children[0].Value.([]byte)
				if len(ipTypeByte) != 2 {
					belogs.Error("extractRoaOidImpl():error ipTypeByte:  bytes:" +
						convert.Bytes2String(roaIPAddressFamilyPack.Children[0].Bytes()) + "  certFile " + certFile)
					return errors.New("error ipTypeByte")
				}
				if ipTypeByte[1] == ipv4 {
					ipType = ipv4
				} else if ipTypeByte[1] == ipv6 {
					ipType = ipv6
				} else {
					belogs.Error("error iptype")
					return errors.New("error iptype")
				}

				//roaIPAddresses := make([]RoaIPAddress, 0)
				for _, roaIPAddressBlock := range roaIPAddressFamilyPack.Children[1].Children {
					roaIpAddressModel := model.RoaIpAddressModel{}
					roaIpAddressModel.AddressFamily = uint64(ipType)

					PrintPacketString("roaIPAddressBlock", roaIPAddressBlock, true, false)

					//roaIPAddress := RoaIPAddress{}
					if len(roaIPAddressBlock.Children) == 0 {
						belogs.Error("extractRoaOidImpl(): cannot found roaIPAddressBlock.Children:" + certFile)
						continue
					}
					ipAddressBlock := roaIPAddressBlock.Children[0]
					PrintPacketString("ipAddressBlock", ipAddressBlock, true, false)
					err = decodeAddressPrefix(ipAddressBlock, ipType)
					if err != nil {
						belogs.Error(err.Error() + ":" + certFile + " decodeAddressPrefix ")
						continue
					}
					roaIpAddressModel.AddressPrefix = ipAddressBlock.Value.(string)
					roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, err =
						iputil.AddressPrefixToHexRange(roaIpAddressModel.AddressPrefix, int(roaIpAddressModel.AddressFamily))
					addressPrefix, err := iputil.FillAddressPrefixWithZero(roaIpAddressModel.AddressPrefix,
						iputil.GetIpType(roaIpAddressModel.AddressPrefix))
					if err != nil {
						belogs.Error("extractRoaOidImpl():FillAddressWithZero err:",
							"addressprefix is "+roaIpAddressModel.AddressPrefix, err)
						return err
					}
					roaIpAddressModel.AddressPrefixRange = jsonutil.MarshalJson(addressPrefix)

					if len(roaIPAddressBlock.Children) > 1 {
						roaIpAddressModel.MaxLength, err = convert.Interface2Uint64(roaIPAddressBlock.Children[1].Value)
						if err != nil {
							belogs.Error(err.Error() + ":" + certFile + "  bytes:" + convert.Bytes2String(roaIPAddressBlock.Children[1].Bytes()))
							continue
						}
					}
					belogs.Debug("roaIPAddress", roaIpAddressModel)

					roaIpAddressModels = append(roaIpAddressModels, roaIpAddressModel)
					belogs.Debug("roaIPAddressFamilies", roaIpAddressModels)
				}
			}
		}
		roaModel.RoaIpAddressModels = roaIpAddressModels
		belogs.Debug("routeOriginAttestation.IPAddrBlocks", roaModel.RoaIpAddressModels)
	}
	belogs.Debug("ExtractRoaOid(): roaModel", roaModel)
	return nil
}

// if decode packet fail ,so try again using OID to decode
func reExtractRoaOid(fileByte []byte, certFile string, roaModel *model.RoaModel) (err error) {

	/*
		may be more one level
		OBJECT IDENTIFIER 1.2.840.113549.1.9.16.1.24 routeOriginAttest (S/MIME Content Types)
		 [0] (1 elem)   //em0
		          OCTET STRING (1 elem)   //oct0
		            OCTET STRING (1 elem) //oct1
		              SEQUENCE (2 elem)		//seq0
		                INTEGER 20015		//asid
		                SEQUENCE (1 elem)	//seq1
		                  SEQUENCE (2 elem)  //seq2
		                    OCTET STRING (2 byte) 0001   //type0
		                    SEQUENCE (32 elem)   //seq3
		                      SEQUENCE (2 elem)   //seq4.1
		                        BIT STRING (24 bit) 110010000100011111000000  //ipaddr
		                        INTEGER 24									  //maxlen
		                      SEQUENCE (2 elem)
		                        BIT STRING (24 bit) 110010000100011111000001
		                        INTEGER 24

	*/
	belogs.Debug("reExtractRoaOid():len(fileByte): ", len(fileByte))
	pos0 := bytes.Index(fileByte, oidRoaKeyByte)
	var datapos uint64 = uint64(pos0)
	var datalen uint64 = uint64(0)
	belogs.Debug("reExtractRoaOid():enum0 pos:", datapos, "  datalen:", datalen)
	if datapos <= 0 {
		return errors.New("not found " + oidManifestKey)
	}
	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	enum0 := fileByte[int(datapos)+len(oidRoaKeyByte):]
	//logs.LogDebugBytes(("reExtractRoaOid():enum0:", enum0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(enum0)
	belogs.Debug("reExtractRoaOid():enum0 pos:", datapos, "  datalen:", datalen)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	oct0 := enum0[datapos:]
	//logs.LogDebugBytes(("reExtractRoaOid():oct0:", oct0)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(oct0)
	belogs.Debug("reExtractRoaOid():oct0 pos:", datapos, "  datalen:", datalen)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	oct1 := oct0[datapos:]
	//logs.LogDebugBytes(("reExtractRoaOid():oct1:", oct1)
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(oct1)
	belogs.Debug("reExtractRoaOid():oct1 pos:", datapos, "   datalen:", datalen)

	//avoid error of 0x00, 0x00, so it is not limit datalen, and will include all data
	seq0 := oct1[datapos:]
	//logs.LogDebugBytes(("reExtractRoaOid():seq0:", seq0)
	seq0Class := seq0[0]
	// if it is not sequence 0x30, then tray a new level
	if seq0Class != byte(0x30) {

		datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(seq0)
		belogs.Debug("reExtractRoaOid():again datalen:", datalen, "    datapos:", datapos)
		seq0 = seq0[int(datapos):]
		//logs.LogDebugBytes(("reExtractRoaOid():again seq0:", seq0)

	} else {

	}
	//calc length of seq, and substring
	datalen, datapos, _ = util.DecodeFiniteAndInfiniteLen(seq0)
	belogs.Debug("reExtractRoaOid():true datalen:", datalen, "    datapos:", datapos)
	seq0 = seq0[:datapos+datalen]

	//donot care about 0x00 0x00, using datalen to get the real length
	//logs.LogDebugBytes(("reExtractRoaOid():true seq0:", seq0)

	pack := DecodePacket(seq0)

	return extractRoaOidImpl(pack, certFile, roaModel)

}
