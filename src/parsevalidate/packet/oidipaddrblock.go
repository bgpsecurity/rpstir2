package packet

import (
	"errors"

	belogs "github.com/astaxie/beego/logs"

	"model"
	"parsevalidate/util"
)

func ExtractIpAddrBlockOid(oidPackets *[]OidPacket) (cerIpAddressModel model.CerIpAddressModel, err error) {
	cerIpAddressModel.CerIpAddresses = make([]model.CerIpAddress, 0)
	var ipType int
	for _, oidPacket := range *oidPackets {
		if oidPacket.Oid == oidIpAddressKey {
			if len(oidPacket.ParentPacket.Children) > 1 {
				critical := oidPacket.ParentPacket.Children[1]
				PrintPacketString("critical", critical, true, false)
				cerIpAddressModel.Critical = util.DecodeBool(critical.Bytes()[0])

				extnValue := oidPacket.ParentPacket.Children[2]
				if len(extnValue.Children) > 0 {
					for _, IpAddressBlocks := range extnValue.Children {
						if len(IpAddressBlocks.Children) > 0 {
							for _, IPAddressFamily := range IpAddressBlocks.Children {

								belogs.Debug("ExtractIpAddrBlockOid():len(IPAddressFamily.Children)", len(IPAddressFamily.Children))
								PrintPacketString("IPAddressFamily", IPAddressFamily, true, false)

								if len(IPAddressFamily.Children) > 0 {
									addressFamily := IPAddressFamily.Children[0]
									PrintPacketString("addressFamily", addressFamily, true, false)

									addressFamilyBytes := addressFamily.Value.([]byte)
									if addressFamilyBytes[1] == ipv4 {
										ipType = ipv4
									} else if addressFamilyBytes[1] == ipv6 {
										ipType = ipv6
									} else {
										belogs.Error("error iptype")
										return cerIpAddressModel, errors.New("error iptype")
									}
									//logs.LogDebugBytes((fmt.Sprintf("addressFamilyBytes: iptype: %d ", ipType), addressFamilyBytes)
									// may be NULL, or be value
									if len(IPAddressFamily.Children) > 1 {
										IPAddressChoice := IPAddressFamily.Children[1]
										PrintPacketString("IPAddressChoice", IPAddressChoice, true, false)
										if len(IPAddressChoice.Children) > 0 {
											for _, addressesOrRangesPacket := range IPAddressChoice.Children {
												PrintPacketString("addressesOrRanges", addressesOrRangesPacket, true, false)
												belogs.Debug("addressesOrRanges: len: ", len(addressesOrRangesPacket.Children))

												cerIpAddress := model.CerIpAddress{}

												if len(addressesOrRangesPacket.Children) > 0 {

													min := addressesOrRangesPacket.Children[0]
													max := addressesOrRangesPacket.Children[1]
													decodeAddressPrefix(min, ipType)
													decodeAddressPrefix(max, ipType)
													PrintPacketString("Range min", min, true, false)
													PrintPacketString("Range max", max, true, false)

													cerIpAddress.Min = min.Value.(string)
													cerIpAddress.Max = max.Value.(string)

												} else {
													decodeAddressPrefix(addressesOrRangesPacket, ipType)
													PrintPacketString("addresses", addressesOrRangesPacket, true, false)

													cerIpAddress.AddressPrefix = addressesOrRangesPacket.Value.(string)
												}
												cerIpAddressModel.CerIpAddresses = append(cerIpAddressModel.CerIpAddresses, cerIpAddress)
											}
										} else {

											if TagNULL == IPAddressChoice.Tag {
												belogs.Debug("IPAddressChoice.Tag is NULL  ", IPAddressChoice.Tag)
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
	return cerIpAddressModel, nil
}
