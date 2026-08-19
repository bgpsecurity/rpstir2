package parsevalidateasn1

import (
	"fmt"
	"testing"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

/*
	func TestParseToAddressPrefix(t *testing.T) {
		//03 04 04 0a0020            addressPrefix 10.0.32/20
		data := []byte{0x03, 0x04, 0x04, 0x0a, 0x00, 0x20}
		address, err := ParseToAddressPrefix(data[2:], 1)
		fmt.Println("\n---------should 10.0.32/20:", address, err)

		//03 04 00 0a0040            addressPrefix 10.0.64/24
		data = []byte{0x03, 0x04, 0x00, 0x0a, 0x00, 0x40}
		address, err = ParseToAddressPrefix(data[2:], 1)
		fmt.Println("\n---------should 10.0.64/24:", address, err)

		//03 03 00 0a01              addressPrefix    10.1/16
		data = []byte{0x03, 0x03, 0x00, 0x0a, 0x01}
		address, err = ParseToAddressPrefix(data[2:], 1)
		fmt.Println("\n---------should 10.1/16:", address, err)

		//30 0c                      addressRange {
		//03 04 04 0a0230           min        10.2.48.0
		//03 04 00 0a0240           max        10.2.64.255
		data = []byte{0x30, 0x0c, 0x03, 0x04, 0x04, 0x0a, 0x02, 0x30, 0x03, 0x04, 0x00, 0x0a, 0x02, 0x40}
		min, max, err := ParseToAddressMinMax(data, 1)
		fmt.Println("\n---------should 10.2.48.0,10.2.64.255:", min, max, err)
	}

	func TestParseToIpNet(t *testing.T) {
		hexStr := `002001067C208C`
		by, err := hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		ipNet, err := ParseToIpNet(by, 2)
		fmt.Println(jsonutil.MarshalJson(ipNet), err)
		addressPrefix, err := ParseToAddressPrefix(by, 2)
		fmt.Println("addressPrefix:", addressPrefix, err)
		newAdd, err := iputil.TrimAddressPrefixZero(addressPrefix, 2)
		fmt.Println("newAdd:", newAdd, err)

		hexStr = `074E8200`
		by, err = hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		ipNet, err = ParseToIpNet(by, 1)
		ones, bits := ipNet.Mask.Size()
		fmt.Println(jsonutil.MarshalJson(ipNet), ones, bits, err)
		addressPrefix, err = ParseToAddressPrefix(by, 1)
		fmt.Println("addressPrefix:", addressPrefix, err)
		newAdd, err = iputil.TrimAddressPrefixZero(addressPrefix, 1)
		fmt.Println("newAdd:", newAdd, err)
	}

	func TestParseToAddressMinMax(t *testing.T) {
		hexStr := `300e030500c0a80001030500c0a80003`
		by, err := hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		min, max, err := ParseToAddressMinMax(by, 1)
		fmt.Println(min, max, err)
	}

	func TestParseToIpAddressBlock(t *testing.T) {
		// 00Z.cer
		//IPv4:   103.121.40.0/22   IPv6:	2403:63c0::/32
		hexStr := `301D300C040200013006030402677928300D040200023007030500240363C0`
		by, err := hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		ipAddrBlocks, err := ParseToIpAddressBlocks(by)
		fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
		fmt.Println("-------------------------")

		// 75414d.cer
		  IPv4:
		  143.137.108.0/22
		  168.181.76.0/22
		  170.150.160.0/22
		  170.244.120.0/22
		  170.245.184.0/22
		  186.194.140.0/22
		  200.53.128.0/18
		  200.57.80.0/20
		  200.77.224.0/20
		  201.159.128.0/20
		  201.175.0.0-201.175.47.255
		IPv6:
		  2001:1270::/32

		hexStr = `3060304F0402000130490304028F896C030402A8B54C030402AA96A0030402AAF478030402AAF5B8030402BAC28C030406C83580030404C83950030404C84DE0030404C99F80300B030300C9AF030404C9AF20300D04020002300703050020011270`
		by, err = hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		ipAddrBlocks, err = ParseToIpAddressBlocks(by)
		fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
		fmt.Println("-------------------------")

			// ipv4+ipv6 range
			hexStr = `3031302004020001301A0304022D40B8030402671BC8300C03040067F5A503040067F5A6300D04020002300703050024077900`
			by, err = hex.DecodeString(hexStr)
			fmt.Println(convert.PrintBytesOneLine(by), err)
			ipAddrBlocks, err = ParseToIpAddressBlocks(by)
			fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
			fmt.Println("-------------------------")

			// only ipv6
			hexStr = `300F300D04020002300703050028041F54`
			by, err = hex.DecodeString(hexStr)
			fmt.Println(convert.PrintBytesOneLine(by), err)
			ipAddrBlocks, err = ParseToIpAddressBlocks(by)
			fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)
			fmt.Println("-------------------------")

			// from sig
			hexStr = `3010300E04010230090307002001067C208C` // error: 040102 --> 04020002
			//hexStr = `3011300F0402000230090307002001067C208C`
			by, err = hex.DecodeString(hexStr)
			fmt.Println(convert.PrintBytesOneLine(by), err)
			ipAddrBlocks, err = ParseToIpAddressBlocks(by)
			fmt.Println(jsonutil.MarshalJson(ipAddrBlocks), err)

}

	func TestParseToAsBlocks(t *testing.T) {
		hexStr := `3009A00730050203020DC6`
		by, err := hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		as, err := ParseToAsBlocks(by)
		fmt.Println(jsonutil.MarshalJson(as), err)
	}

	func TestParseToFileAndHashs(t *testing.T) {
		hexStr := `3077303416106234325F697076365F6C6F612E706E6704209516DD64BE7C1725B9FCA117120E58E8D842A5206873399B3DDFFC91C4B6ACF0303F161B6234325F736572766963655F646566696E6974696F6E2E6A736F6E04200AE1394722005CD92F4C6AA024D5D6B3E2E67D629F11720D9478A633A117A1C7`
		by, err := hex.DecodeString(hexStr)
		fmt.Println(convert.PrintBytesOneLine(by), err)
		fs, err := ParseToFileAndHashs(by)
		fmt.Println(jsonutil.MarshalJson(fs), err)
	}
*/
func TestParseCerFile(t *testing.T) {
	path := `G:\Download\cert\cer\`
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	files, _ := osutil.GetFilesInDir(path, m)
	fmt.Println("len(files):", len(files))
	start := time.Now()
	for _, file := range files {
		fileModel := &model.FileModel{
			FilePathName:     path + file,
			TempFilePathName: path + file,
		}

		fmt.Println("fileModel:", fileModel)
		cerModel := new(model.CerModel)
		stateModel := new(model.StateModel)
		err := ParseCerModelByAsn1(fileModel, false, cerModel, stateModel)
		if err != nil {
			fmt.Println(file, err)
			return
		}
		fmt.Println(file, "\n", jsonutil.MarshalJson(cerModel)+"\n\n\n")
	}
	fmt.Println("end len(files):", len(files), "  time(s):", time.Since(start))
}
