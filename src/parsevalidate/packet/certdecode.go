package packet

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	belogs "github.com/astaxie/beego/logs"

	"parsevalidate/util"
)

//found hierarchy by "0x00 0x00"
func DecodePacket(data []byte) *Packet {
	topHierarchyFor00 := util.GetTopHierarchyFor00(data)
	p, _, err := decodePacketImpl(data, 0, topHierarchyFor00)
	if err != nil {
		belogs.Error("DecodePacket", err)
		return nil
	}
	return p
}

func decodePacketImpl(data []byte, hierarchyFor00 int, topHierarchyFor00 int) (*Packet, []byte, error) {
	belogs.Debug("decodePacketImpl: enter len(data):", len(data), "  hierarchyFor00:", hierarchyFor00, "   topHierarchyFor00:", topHierarchyFor00)
	//logs.LogDebugBytes("decodePacketImpl: enter data: ", data)

	if len(data) < 2 {
		return nil, nil, errors.New("data is empty")
	}
	// may have 0x00, 0x00, should be trimed
	data = util.TrimPrefix00(data)
	if len(data) == 0 {
		return nil, nil, nil
	}
	var err error
	p := new(Packet)
	p.ClassType = data[0] & ClassBitmask
	p.TagType = data[0] & TypeBitmask
	p.Tag = data[0] & TagBitmask
	belogs.Debug("decodePacketImpl: p.ClassType=%d, p.TagType=%d, p.Tag=%d ",
		p.ClassType, p.TagType, p.Tag)

	datalen := util.DecodeInteger(data[1:2])
	belogs.Debug("decodePacketImpl datalen in data[1:2]: ", datalen)

	//may have problems to calc length , see
	//https://blog.csdn.net/sever2012/article/details/7698297
	//http://javadoc.iaik.tugraz.at/iaik_jce/current/iaik/asn1/DerCoder.html
	//https://en.wikipedia.org/wiki/X.690

	datapos := uint64(2)
	if datalen == 0x80 {
		//Length is not specified. the "0x00 0x00" shows end
		// 0x30 80, 0xa0 80,  0x24 80,
		datapos = uint64(2)
		//lastpos start at head, include 0xtype and 0x80, and not include the last 2 bytes(0x00 0x00)
		lastpos, err := util.IndexEndOfBytes(data, p.TagType, hierarchyFor00, topHierarchyFor00)
		belogs.Debug("decodePacketImpl datalen == 0x80: lastpos: ", lastpos)
		if err != nil {
			return nil, nil, err
		}
		//delete 2bytes at head： first byte is type, second byte is 0x80
		datalen = uint64(lastpos - 1 - 1)
		belogs.Debug("decodePacketImpl datalen == 0x80: datapos: ", datapos, ",   datalen: ", datalen, ",   len(data): ", len(data))

	} else if datalen != 0x80 && datalen&128 != 0 {
		//FiniteLen will set the length of value
		/* */
		datalen, datapos, err = util.DecodeFiniteLen(data)
		if err != nil {
			return nil, nil, nil
		}

	}

	// if lastindexof 0x00 0x00 fail ,then try again, from indexof 0x00 0x00
	if datapos+datalen > uint64(len(data)) {
		belogs.Debug("decodePacketImpl() datapos+datalen > uint64(len(data) ", datapos, datalen, len(data))
		lastpos, err := util.IndexEndOfBytes(data, p.TagType, hierarchyFor00, topHierarchyFor00)
		datalen = uint64(lastpos - 1 - 1)
		belogs.Debug("decodePacketImpl IndexEndOfBytes again  datapos: ", datapos, ",   datalen: ", datalen, ",   len(data): ", uint64(len(data)))

		if err != nil || datapos+datalen > uint64(len(data)) {
			return nil, nil, nil // errors.New("data is less than datapos+datalen")
		}
	}
	valueData := data[datapos : datapos+datalen]
	valueDataTmp := util.TrimPrefix00(valueData)
	//logs.LogDebugBytes(("valueDataTmp", valueDataTmp)
	if len(valueDataTmp) == 0 {
		return nil, nil, nil
	}

	belogs.Debug("decodePacketImpl():  [datapos : datapos+datalen], len(valueData) : ", datapos, datalen, len(valueData))
	//logs.LogDebugBytes("decodePacketImpl valueData bytes:", valueData)

	p.Data = new(bytes.Buffer)
	p.Children = make([]*Packet, 0, 2)
	p.Value = nil

	/*
		https://blog.csdn.net/liaowenfeng/article/details/8777595
		ASN.1, left 2bit, is type
			    0	1	type
				0	0	Universal
				0	1	Application
				1	0	Context Specific
				1	1	Private
	*/

	if p.TagType == TypeConstructed {

		belogs.Debug("after p.TagType == TypeConstructed, len(valueData): ", len(valueData))
		for len(valueData) != 0 {
			var child *Packet
			var err error
			child, valueData, err = decodePacketImpl(valueData, hierarchyFor00+1, topHierarchyFor00)
			if err != nil {
				belogs.Debug("decodePacketImpl error:", err)
				return nil, nil, nil // nil, nil, err //no return error
			}
			if child != nil {
				p.AppendChild(child)
			}
		}
	} else if p.ClassType == ClassUniversal {

		//logs.LogDebugBytes(("after p.ClassType == ClassUniversal:", data[datapos:datapos+datalen])
		p.Data.Write(data[datapos : datapos+datalen])
		//logs.LogDebugBytes(("after p.Data.Write(data[datapos : datapos+datalen]) :", p.Bytes())
		switch p.Tag {
		case TagEOC:
		case TagBoolean:
			val := util.DecodeInteger(valueData)
			p.Value = val != 0
		case TagInteger:
			p.Value = util.DecodeInteger(valueData)
		case TagBitString:
			p.Value = (valueData)
		case TagOctetString:
			//p.Value = DecodeString(valueData)
			// OctetString may have children, should get the first bit of child
			haveChild := false
			if len(valueData) > 0 {
				//valueDataSaved := data[datapos : datapos+datalen]
				childTagType := valueData[0] & TypeBitmask
				belogs.Debug("decodePacketImpl: childTagType %d, (%d);   value_date[0]=%d, (%d)\n", childTagType, TypeConstructed, valueData[0], TagBitmask)
				//logs.LogDebugBytes(("before childTagType is TypeConstructed:", valueData)
				if int(childTagType) == TypeConstructed {
					//var child *Packet
					//logs.LogDebugBytes(("TagOctetString before:", valueData)
					child, _, err := decodePacketImpl(valueData, hierarchyFor00+1, topHierarchyFor00)
					//logs.LogDebugBytes(("TagOctetString before err==nil :", valueDataSaved)
					// if here has error, it means it is not child, it is just string. So, it is not return error, just return string
					if err == nil && child != nil {
						//return nil, nil, err
						//logs.LogDebugBytes(("TagOctetString after decodePacketImpl :", p.Bytes())
						// clear old bytes, set new bytes of child
						p.Data.Reset()
						p.AppendChild(child)
						haveChild = true
					} else {

					}

				}

			}
			//if there is no children, will set bytes
			if !haveChild {
				p.Value = valueData
				/*
					var buf bytes.Buffer
					enc := gob.NewEncoder(&buf)
					err := enc.Encode(key)
				*/
			}
			break
		case TagNULL:
			p.Value = nil
		case TagObjectIdentifier:
			p.Value = util.DecodeOid(valueData)
			//fmt.Println(p.Value.(string))
		case TagObjectDescriptor:
		case TagExternal:
		case TagRealFloat:
		case TagEnumerated:
			p.Value = util.DecodeInteger(valueData)
		case TagEmbeddedPDV:
		case TagUTF8String:
			p.Value = util.DecodeUTF8String(valueData)
		case TagRelativeOID:
		case TagSequence:
		case TagSet:
		case TagNumericString:
		case TagPrintableString:
			p.Value = util.DecodeString(valueData)
		case TagT61String:
		case TagVideotexString:
		case TagIA5String:
			p.Value = util.DecodeIA5String(valueData)
		case TagUTCTime:
			p.Value, err = util.DecodeUTCTime(valueData)
		case TagGeneralizedTime:
			p.Value, err = util.DecodeGeneralizedTime(valueData)
		case TagGraphicString:
		case TagVisibleString:
		case TagGeneralString:
		case TagUniversalString:
		case TagCharacterString:
		case TagBMPString:
		//private
		case TagAsNum:
			p.Value = valueData
		case TagRdi:
			p.Value = valueData
		}
	} else {
		p.Data.Write(data[datapos : datapos+datalen])
	}
	//logs.LogDebugBytes(("decodePacketImpl: end switch", data[datapos+datalen:])
	return p, data[datapos+datalen:], err
}

func (p *Packet) DataLength() uint64 {
	return uint64(p.Data.Len())
}

func (p *Packet) Bytes() []byte {
	var out bytes.Buffer
	out.Write([]byte{p.ClassType | p.TagType | p.Tag})
	packetLength := util.EncodeInteger(p.DataLength())
	if p.DataLength() > 127 || len(packetLength) > 1 {
		out.Write([]byte{byte(len(packetLength) | 128)})
		out.Write(packetLength)
	} else {
		out.Write(packetLength)
	}
	out.Write(p.Data.Bytes())
	return out.Bytes()
}

func (p *Packet) AppendChild(child *Packet) {
	p.Data.Write(child.Bytes())
	if len(p.Children) == cap(p.Children) {
		newChildren := make([]*Packet, cap(p.Children)*2)
		copy(newChildren, p.Children)
		p.Children = newChildren[0:len(p.Children)]
	}
	p.Children = p.Children[0 : len(p.Children)+1]
	p.Children[len(p.Children)-1] = child
}

func TransformPacket(p *Packet, oidPackets *[]OidPacket) {

	for i := range p.Children {

		p.Children[i].Parent = p
		//fmt.Println(p.Children[i].Tag, TagObjectIdentifier)
		if p.Children[i].Tag == TagObjectIdentifier {
			oidPacket := OidPacket{}
			//fmt.Printf("%s%s(%s, %s, %s) Len=%d %q\n", indent_str, description, class_str, tagtype_str, tag_str, p.Data.Len(), value)
			//fmt.Println(p.Children[i].Value.(string))
			oid := fmt.Sprint(p.Children[i].Value)
			oidPacket.Oid = oid
			//fmt.Println("addParent():oid:", oid)
			oidPacket.ParentPacket = p
			oidPacket.OidPacket = p.Children[i]

			belogs.Debug("TransformPacket(): ", oidPacket)
			//logs.LogDebugBytes(("TransformPacket():oidPacket:", p.Bytes())

			(*oidPackets) = append((*oidPackets), oidPacket)
		}
		TransformPacket(p.Children[i], oidPackets)
	}
}

func decodeAddressPrefix(addressPrefixPacket *Packet, ipType int) error {
	addressPrefix := addressPrefixPacket.Bytes()
	//logs.LogDebugBytes(("decodeAddressPrefix():oidPacket:", addressPrefix)
	if len(addressPrefix) < 4 {
		belogs.Error("decodeAddressPrefix():len(addressPrefix)<3", addressPrefix)
		return fmt.Errorf("addressPrefix length is error: %d", len(addressPrefix))
	}
	addressShouldLen, _ := strconv.Atoi(fmt.Sprintf("%d", addressPrefix[1]))
	unusedBitLen, _ := strconv.Atoi(fmt.Sprintf("%d", addressPrefix[2]))

	address := addressPrefix[4:]
	ipAddress := ""

	if ipType == ipv4 {
		// ipv4 ipaddress prefx
		prefix := 8*(addressShouldLen-1) - unusedBitLen
		belogs.Debug("prefix := 8*(addressShouldLen-1) - unusedBitLen:  %d := 8 *(%d-1)-  %d \r\n",
			prefix, addressShouldLen, unusedBitLen)

		ipv4Address := ""
		for i := 0; i < len(address); i++ {
			ipv4Address += fmt.Sprintf("%d", address[i])
			if i < len(address)-1 {
				ipv4Address += "."
			}
		}
		ipv4Address += "/" + fmt.Sprintf("%d", prefix)
		ipAddress = ipv4Address
		belogs.Debug("ipv4Address:%s", ipv4Address)

	} else if ipType == ipv6 {
		// ipv6 ipaddress prefx
		prefix := 8*(addressShouldLen-1) - unusedBitLen
		belogs.Debug("prefix :=  8*(addressShouldLen-1) - unusedBitLen:  %d := 8 *(%d-1)-  %d \r\n",
			prefix, addressShouldLen, unusedBitLen)

		//printBytes("address:", address)

		ipv6Address := ""
		for i := 0; i < len(address); i++ {
			ipv6Address += fmt.Sprintf("%02x", address[i])
			if i%2 == 1 && i < len(address)-1 {
				ipv6Address += ":"
			}
		}
		//Complement digit
		if len(address)%2 == 1 {
			ipv6Address += "00"
		}
		ipv6Address += "/" + fmt.Sprintf("%d", prefix)
		ipAddress = ipv6Address
		belogs.Debug("ipv6Address:%s", ipv6Address)

	}
	addressPrefixPacket.Value = ipAddress
	return nil
}

func PrintPacketString(name string, p *Packet, printBytes bool, printChild bool) {

	classStr := ClassMap[p.ClassType]
	tagtypeStr := TypeMap[p.TagType]
	tagStr := fmt.Sprintf("0x%02X", p.Tag)

	if p.ClassType == ClassUniversal {
		tagStr = TagMap[p.Tag]
	}

	value := fmt.Sprint(p.Value)
	description := ""
	if p.Description != "" {
		description = p.Description + ": "
	}

	belogs.Debug("PrintPacketString():%s  %s(%s, %s, %s) Len=%v %q", name, description, classStr, tagtypeStr, tagStr, p.Data.Len(), value)

	if printBytes {
		//logs.LogDebugBytes(("", p.Bytes())
	}
	if printChild {
		for _, child := range p.Children {
			belogs.Debug("[children]-->")
			PrintPacketString(name+" --> children ", child, printBytes, printChild)
		}
	}
}

func PrintOidPacket(oidPackets *[]OidPacket) {

	belogs.Debug("all oidPacket size:%d", len(*oidPackets))
	for _, oidPacket := range *oidPackets {

		belogs.Debug(oidPacket.Oid)
		//logs.LogDebugBytes(("oid parent bytes:", oidPacket.ParentPacket.Bytes())
		//logs.LogDebugBytes(("oid self bytes:", oidPacket.OidPacket.Bytes())
		belogs.Debug("")
	}
}
