package util

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	belogs "github.com/astaxie/beego/logs"
)

var oid = map[string]string{
	"2.5.4.3":                    "CN",
	"2.5.4.4":                    "SN",
	"2.5.4.5":                    "serialNumber",
	"2.5.4.6":                    "C",
	"2.5.4.7":                    "L",
	"2.5.4.8":                    "ST",
	"2.5.4.9":                    "streetAddress",
	"2.5.4.10":                   "O",
	"2.5.4.11":                   "OU",
	"2.5.4.12":                   "title",
	"2.5.4.17":                   "postalCode",
	"2.5.4.42":                   "GN",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generationQualifier",
	"2.5.4.46":                   "dnQualifier",
	"2.5.4.65":                   "pseudonym",
	"0.9.2342.19200300.100.1.25": "DC",
	"1.2.840.113549.1.9.1":       "emailAddress",
	"0.9.2342.19200300.100.1.1":  "userid",
	"2.5.29.20":                  "CRL Number",
}

func GetDNFromName(namespace pkix.Name, sep string) (string, error) {
	return GetDNFromRDNSeq(namespace.ToRDNSequence(), sep)
}

func GetDNFromRDNSeq(rdns pkix.RDNSequence, sep string) (string, error) {
	subject := []string{}
	for _, s := range rdns {
		for _, i := range s {
			if v, ok := i.Value.(string); ok {
				if name, ok := oid[i.Type.String()]; ok {
					// <oid name>=<value>
					subject = append(subject, fmt.Sprintf("%s=%s", name, v))
				} else {
					// <oid>=<value> if no <oid name> is found
					subject = append(subject, fmt.Sprintf("%s=%s", i.Type.String(), v))
				}
			} else {
				// <oid>=<value in default format> if value is not string
				subject = append(subject, fmt.Sprintf("%s=%v", i.Type.String, v))
			}
		}
	}
	return strings.Join(subject, sep), nil
}

func DecodeString(data []byte) (ret string) {
	for _, c := range data {
		ret += fmt.Sprintf("%c", c)
	}
	return
}
func DecodeUTF8String(data []byte) (ret string) {

	return string(data)
}

func DecodeIA5String(data []byte) (ret string) {
	return string(data)
}

func DecodeBool(data byte) (ret bool) {
	if data == 0x00 {
		return false
	}
	return true
}

// UTC is short Year, 2 nums
func DecodeUTCTime(data []byte) (ret string, err error) {
	if len(data) < 13 {
		return "", errors.New("DecodeUTCTime fail")
	}
	year := "20" + string(data[0:2])
	month := string(data[2:4])
	day := string(data[4:6])
	hour := string(data[6:8])
	minute := string(data[8:10])
	second := string(data[10:12])
	z := string(data[12])
	return year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second + z, nil
}

// Generalized Long Year, 4 nums
func DecodeGeneralizedTime(data []byte) (ret string, err error) {
	if len(data) < 15 {
		return "", errors.New("DecodeGeneralizedTime fail")
	}
	year := string(data[0:4])
	month := string(data[4:6])
	day := string(data[6:8])
	hour := string(data[8:10])
	minute := string(data[10:12])
	second := string(data[12:14])
	z := string(data[14])
	return year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second + z, nil
}

func DecodeOid(data []byte) (ret string) {

	oids := make([]uint32, len(data)+2)
	//the first byte using: first_arc* 40+second_arc
	//the later , when highest bit is 1, will add to next to calc
	// https://msdn.microsoft.com/en-us/library/windows/desktop/bb540809(v=vs.85).aspx
	f := uint32(data[0])
	if f < 80 {
		oids[0] = f / 40
		oids[1] = f % 40
	} else {
		oids[0] = 2
		oids[1] = f - 80
	}
	var tmp uint32
	for i := 2; i <= len(data); i++ {
		f = uint32(data[i-1])
		//	fmt.Printf("f:0x%x\r\n", f)
		if f >= 0x80 {
			//		fmt.Printf("tmp<<8:0x%x +   (f&0x7f)0x%x\r\n", tmp<<8, (f & 0x7f))
			tmp = tmp<<7 + (f & 0x7f)
			//		fmt.Printf("tmp:0x%x\r\n", tmp)
		} else {
			oids[i] = tmp<<7 + (f & 0x7f)
			//		fmt.Printf("oids[i]:0x%x\r\n", oids[i])
			tmp = 0
		}
	}
	var buffer bytes.Buffer
	for i := 0; i < len(oids); i++ {
		if oids[i] == 0 {
			continue
		}
		buffer.WriteString(fmt.Sprint(oids[i]) + ".")
	}
	belogs.Debug("DecodeOid(): oid:", buffer.String()[0:len(buffer.String())-1])
	return buffer.String()[0 : len(buffer.String())-1]
}

// return byte directly
func DecodeBytes(data []byte) (ret string) {
	for _, b := range data {
		ret += fmt.Sprintf("%02x ", b)
	}
	return ret
}

func EncodeInteger(val uint64) []byte {
	var out bytes.Buffer
	found := false
	shift := uint(56)
	mask := uint64(0xFF00000000000000)
	for mask > 0 {
		if !found && (val&mask != 0) {
			found = true
		}
		if found || (shift == 0) {
			out.Write([]byte{byte((val & mask) >> shift)})
		}
		shift -= 8
		mask = mask >> 8
	}
	return out.Bytes()
}

func ReadFileAndDecodeBase64(file string) (fileByte []byte, fileDecodeBase64Byte []byte, err error) {
	fileByte, err = ioutil.ReadFile(file)
	if err != nil {
		belogs.Error("ReadFile():err:", file, err)
		return nil, nil, err
	}
	fileDecodeBase64Byte, err = DecodeBase64(fileByte)
	if err != nil {
		belogs.Error("ReadFile():DecodeBase64 err:", file, err)
		return nil, nil, err
	}
	return fileByte, fileDecodeBase64Byte, nil
}

func DecodeBase64(oldBytes []byte) ([]byte, error) {
	isBinary := false

	for _, b := range oldBytes {
		t := int(b)

		if t < 32 && t != 9 && t != 10 && t != 13 {
			isBinary = true
			break
		}
	}

	belogs.Debug("DecodeBase64(): isBinary:", isBinary)
	if isBinary {
		return oldBytes, nil
	}
	txt := string(oldBytes)
	txt = strings.Replace(txt, "-----BEGIN CERTIFICATE-----", "", -1)
	txt = strings.Replace(txt, "-----END CERTIFICATE-----", "", -1)
	txt = strings.Replace(txt, "-", "", -1)
	txt = strings.Replace(txt, " ", "", -1)
	txt = strings.Replace(txt, "\r", "", -1)
	txt = strings.Replace(txt, "\n", "", -1)
	belogs.Debug("DecodeBase64(): txt after Replace: %s", txt)
	newBytes, err := base64.StdEncoding.DecodeString(txt)
	return newBytes, err

}
func DecodeInteger(data []byte) (ret uint64) {
	for _, i := range data {
		ret = ret * 256
		ret = ret + uint64(i)
	}
	return
}

// FiniteLen get length
func DecodeFiniteLen(data []byte) (datalen uint64, datapos uint64, err error) {
	datalen = DecodeInteger(data[1:2])
	datapos = uint64(2)
	if datalen&128 != 0 {
		datalen -= 128
		datapos += datalen
		if 2+datalen >= uint64(len(data)) {
			belogs.Debug("DecodeFiniteLen():data is less than 2+datalen")
			return 0, 0, errors.New("data is less than datalen")
		}
		belogs.Debug("DecodeFiniteLen(): 2+datalen: ", 2+datalen, "   uint64(len(data):", uint64(len(data)))
		//logs.LogDebugBytes(("DecodeFiniteLen(): data[2:2+datalen]:", data[2:2+datalen])

		datalen = DecodeInteger(data[2 : 2+datalen])
		belogs.Debug("DecodeFiniteLen(): datalen in data[2 : 2+datalen]: ", datalen)
	}
	belogs.Debug("DecodeFiniteLen():return datalen: ", datalen, " datapos:", datapos)
	return datalen, datapos, nil

}

// InfiniteLen just care about the 0x00 0x00
func DecodeInfiniteLen(data []byte) (datalen uint64, datapos uint64, err error) {
	endbytes := []byte{0x00, 0x00}
	datalen = uint64(bytes.Index(data, endbytes))
	datapos = uint64(2)
	return datalen, datapos, nil
}

// FiniteLen will get length, but InfiniteLen using 0x00 0x00 to get length
func DecodeFiniteAndInfiniteLen(data []byte) (datalen uint64, datapos uint64, err error) {
	data0Len := data[1]
	belogs.Debug("DecodeFiniteAndInfiniteLen():again seq0Len:", data0Len)
	if data0Len == byte(0x80) {
		datalen, datapos, _ = DecodeInfiniteLen(data)
	} else {
		datalen, datapos, _ = DecodeFiniteLen(data)
	}
	belogs.Debug("DecodeFiniteAndInfiniteLen():datalen:", datalen, " datapos:", datapos)
	return datalen, datapos, nil
}

//found the location of 0x00 0x00
func IndexEndOfBytes(oldb []byte, tagType uint8, hierarchyFor00 int, topHierarchyFor00 int) (int, error) {
	belogs.Debug("IndexEndOfBytes():len(oldb):", len(oldb), "  tagType", tagType,
		"   hierarchyFor00:", hierarchyFor00, "     topHierarchyFor00:", topHierarchyFor00)
	//logs.LogDebugBytes(("IndexEndOfBytes(): oldb:", oldb)
	if len(oldb) <= 2 {
		return -1, errors.New("bytes is too short")
	}
	//0x30 80, 0011 0000
	//0xa0 80, 1010 0000
	//0x24 80, 0010 0100
	var pos int
	endbytes := []byte{0x00, 0x00}
	var TypeConstructed uint8 = 32 // xx1xxxxxb  0011 0010
	var TypeLastIndex byte = 0xa0  // see certpacket
	if oldb[0] == TypeLastIndex || (tagType == TypeConstructed && hierarchyFor00 < topHierarchyFor00) {
		pos = bytes.LastIndex(oldb, endbytes)
		belogs.Debug("IndexEndOfBytes(): LastIndex  pos:", pos)
	} else {
		// may be more 0x00 0x00 are together, found the latest 0x00 0x00
		pos = bytes.Index(oldb, endbytes)
		//logs.LogDebugBytes(("IndexEndOfBytes():Index:", oldb[:pos+2])
		for len(oldb) > 0 &&
			pos > 0 &&
			len(oldb) > pos+2*len(endbytes) &&
			len(oldb) > pos+4 &&
			bytes.Equal(oldb[pos+2:pos+4], endbytes) {
			pos += 2
			belogs.Debug("IndexEndOfBytes():for Index  pos:", pos)
			//logs.LogDebugBytes(("IndexEndOfBytes():for Index:", oldb[:pos+2])
		}

		belogs.Debug("IndexEndOfBytes(): Index  pos:", pos)

	}

	if pos < 0 {
		return len(oldb), nil
	}
	return pos, nil

}

func GetTopHierarchyFor00(oldb []byte) int {
	top := 0
	endbytes := []byte{0x00, 0x00}
	pos := bytes.LastIndex(oldb, endbytes)
	for pos > 0 && len(oldb) == pos+len(endbytes) {
		oldb = oldb[:pos]
		top += 1
		pos = bytes.LastIndex(oldb, endbytes)
	}
	belogs.Debug("GetTopHierarchyFor00(): top:", top)
	return top
}

func TrimSuffix00(oldByte []byte, cerEndIndex int) (b []byte, i int) {
	nullbytes := []byte{0x00, 0x00}
	if bytes.HasSuffix(oldByte, nullbytes) {
		oldByte = oldByte[:len(oldByte)-len(nullbytes)]
		cerEndIndex = cerEndIndex - len(nullbytes)
	}
	return oldByte, cerEndIndex
}

func TrimPrefix00(olddb []byte) []byte {
	if len(olddb) <= 2 {
		return olddb
	}
	//trim the head 0x00 0x00
	nullbytes := []byte{0x00, 0x00}
	for bytes.HasPrefix(olddb, nullbytes) {
		olddb = olddb[len(nullbytes):]
		//logs.LogDebugBytes(("TrimNull(): 0x00 0x00, bytes.HasPrefix olddb:", olddb)
	}
	// trim the head 0x00
	nullbytes = []byte{0x00}
	for bytes.HasPrefix(olddb, nullbytes) {
		olddb = olddb[len(nullbytes):]
		//logs.LogDebugBytes(("TrimNull(): 0x00,  bytes.HasPrefix olddb:", olddb)
	}
	//if the second bit of head is not 0x80, and the tail is 0x00 0x00, the trim the tail 0x00 0x00
	/*
		indefinitebytes := byte(0x80)
		for len(olddb) > 4 && olddb[1] != indefinitebytes && bytes.HasSuffix(olddb, nullbytes) {
			olddb = olddb[:len(olddb)-len(nullbytes)]
			//logs.LogDebugBytes(("TrimNull(): bytes.HasSuffix olddb:", olddb, "\t")
		}
	*/
	return olddb
}

func ExtKeyUsagesToInts(exts []x509.ExtKeyUsage) []int {
	ks := make([]int, 0)
	for _, e := range exts {
		ks = append(ks, int(e))
	}
	return ks
}
