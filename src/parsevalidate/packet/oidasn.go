package packet

import (
	belogs "github.com/astaxie/beego/logs"

	"model"
	"parsevalidate/util"
)

func ExtractASNOid(oidPackets *[]OidPacket) (asnModel model.AsnModel, err error) {
	asnModel.Asns = make([]model.Asn, 0)

	for _, oidPacket := range *oidPackets {

		if oidPacket.Oid == oidASKey {
			if len(oidPacket.ParentPacket.Children) > 1 {
				critical := oidPacket.ParentPacket.Children[1]
				PrintPacketString("critical", critical, true, false)
				asnModel.Critical = util.DecodeBool(critical.Bytes()[0])

				extnValue := oidPacket.ParentPacket.Children[2]
				PrintPacketString("extnValue", extnValue, true, false)
				if len(extnValue.Children) > 0 {
					for _, ASIdentifiers := range extnValue.Children {
						PrintPacketString("ASIdentifiers", ASIdentifiers, true, false)
						if len(ASIdentifiers.Children) > 0 {
							for _, ASIdentifierChoice := range ASIdentifiers.Children {
								PrintPacketString("ASIdentifierChoice", ASIdentifierChoice, true, false)
								if len(ASIdentifierChoice.Children) > 0 {
									for _, asIdsOrRanges := range ASIdentifierChoice.Children {
										PrintPacketString("asIdsOrRanges", asIdsOrRanges, true, false)
										if len(asIdsOrRanges.Children) > 0 {
											for _, ASIdOrRangePacket := range asIdsOrRanges.Children {

												asn := model.NewAsn()
												//there are 2 types: 1 chilren is ASRang, 2 ASId
												if len(ASIdOrRangePacket.Children) > 1 {
													min := ASIdOrRangePacket.Children[0]
													max := ASIdOrRangePacket.Children[1]

													asn.Min = min.Value.(int64)
													asn.Max = max.Value.(int64)

													PrintPacketString("ASNum min", min, true, false)
													PrintPacketString("ASNum max", max, true, false)
												} else {

													asn.Asn = ASIdOrRangePacket.Value.(int64)
													PrintPacketString("ASId", ASIdOrRangePacket, true, false)
												}
												asnModel.Asns = append(asnModel.Asns, asn)
											}
										} else {

											if TagNULL == asIdsOrRanges.Tag {
												belogs.Debug("asIdsOrRanges.Tag is NULL  ", asIdsOrRanges.Tag)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

	}
	return asnModel, nil
}
