package openssl

import (
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cpusoft/goutil/asn1util"
	"github.com/cpusoft/goutil/asn1util/asn1base"
	"github.com/cpusoft/goutil/asn1util/asn1cert"
	"github.com/cpusoft/goutil/asn1util/asn1node"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/datetime"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"

	model "rpstir2-model"
)

type ManifestParse struct {
	ManifestNumber asn1.RawValue         `json:"manifestNumber"`
	ThisUpdate     time.Time             `asn1:"generalized" json:"thisUpdate"`
	NextUpdate     time.Time             `asn1:"generalized" json:"nextUpdate"`
	FileHashAlg    asn1.ObjectIdentifier `json:"fileHashAlg"`
	FileList       []FileAndHashParse    `json:"fileList"`
}
type FileAndHashParse struct {
	File string         `asn1:"ia5" json:"file"`
	Hash asn1.BitString `json:"hash"`
}

type ManifestRawParse struct {
	ManifestNumber asn1.RawValue         `json:"manifestNumber"`
	ThisUpdate     time.Time             `asn1:"generalized" json:"thisUpdate"`
	NextUpdate     time.Time             `asn1:"generalized" json:"nextUpdate"`
	FileHashAlg    asn1.RawValue         `json:"fileHashAlg"`
	FileList       []FileAndHashRawParse `json:"fileList"`
}
type FileAndHashRawParse struct {
	File asn1.RawValue `json:"file"`
	Hash asn1.RawValue `json:"hash"`
}

func ParseMftModelByOpensslResults(results []string, mftModel *model.MftModel) (err error) {
	// get mft info
	var mftHex string
	foundAllMftHex := false
	keyword := "[HEX DUMP]:"
	for i, one := range results {
		if strings.Contains(one, keyword) {
			index := strings.Index(one, keyword)
			mftHex = string(one[index+len(keyword):])
			belogs.Debug("ParseMftModelByOpensslResults(): len(mftHex):", len(mftHex))

			if !strings.Contains(results[i+1], keyword) {
				foundAllMftHex = true
				belogs.Debug("ParseMftModelByOpensslResults(): foundAllMftHex:", foundAllMftHex)
				break
			}
			// one [HEX DUMP] length is 10000, so some mft have many ip, will have many [HEX DUMP],
			// just one loop using foundmftHexHex to break
			for ii := i + 1; ii < len(results); ii++ {
				if strings.Contains(results[ii], keyword) {
					indexii := strings.Index(results[ii], keyword)
					mftHexii := string(results[ii][indexii+len(keyword):])
					mftHex = mftHex + mftHexii
				} else {
					foundAllMftHex = true
					break
				}
			}
		}
		if foundAllMftHex {
			break
		}
	}
	if len(mftHex) == 0 {
		belogs.Error("ParseMftModelByOpensslResults():len(mftHex) == 0")
		return errors.New("not found mft hex")
	}

	mftByte, err := hex.DecodeString(mftHex)
	if err != nil {
		belogs.Error("ParseMftModelByOpensslResults():DecodeString err:", err)
		return err
	}
	mft := ManifestParse{}
	asn1.Unmarshal(mftByte, &mft)
	belogs.Debug("ParseMftModelByOpensslResults(): mft:", jsonutil.MarshalJson(mft))

	if len(mft.FileList) > 0 {
		belogs.Debug("ParseMftModelByOpensslResults():mft.FileList>0 ")

		mftNumber := convert.Bytes2String(mft.ManifestNumber.Bytes)
		belogs.Debug("ParseMftModelByOpensslResults():mftNumber: ", mftNumber)
		mftModel.MftNumber = mftNumber
		mftModel.ThisUpdate = mft.ThisUpdate
		mftModel.NextUpdate = mft.NextUpdate
		mftModel.FileHashAlg = mft.FileHashAlg.String()

		belogs.Debug("ParseMftModelByOpensslResults():mft.FileList: ", jsonutil.MarshalJson(mft.FileList))
		fileHashModels := make([]model.FileHashModel, 0)
		for _, one := range mft.FileList {
			fileHashModel := model.FileHashModel{}
			fileHashModel.File = one.File
			fileHashModel.Hash = convert.Bytes2String(one.Hash.Bytes)
			fileHashModels = append(fileHashModels, fileHashModel)
			belogs.Debug("ParseMftModelByOpensslResults(): fileHashModel: ", jsonutil.MarshalJson(fileHashModel))
		}
		mftModel.FileHashModels = fileHashModels
		belogs.Debug("ParseMftModelByOpensslResults(): mftModel:", jsonutil.MarshalJson(mftModel))
		return nil
	}
	// using raw try again
	mftRaw := ManifestRawParse{}
	asn1.Unmarshal(mftByte, &mftRaw)
	if len(mftRaw.FileList) == 0 {
		belogs.Error("ParseMftModelByOpensslResults():mftRaw.FileList==0: ", jsonutil.MarshalJson(mftRaw))
		return errors.New("parse mft FileList error")
	}
	mftRawNumber := convert.Bytes2String(mftRaw.ManifestNumber.Bytes)
	belogs.Debug("ParseMftModelByOpensslResults():mftRawNumber: ", mftRawNumber)
	mftModel.MftNumber = mftRawNumber
	mftModel.ThisUpdate = mftRaw.ThisUpdate
	mftModel.NextUpdate = mftRaw.NextUpdate
	//mftModel.FileHashAlg = mftRaw.FileHashAlg.String()

	belogs.Debug("ParseMftModelByOpensslResults():mftRaw.FileList: ", jsonutil.MarshalJson(mftRaw.FileList))
	fileHashModels := make([]model.FileHashModel, 0)
	for _, one := range mftRaw.FileList {
		fileHashModel := model.FileHashModel{}
		fileHashModel.File = convert.Bytes2String(one.File.Bytes)
		fileHashModel.Hash = convert.Bytes2String(one.Hash.Bytes)
		fileHashModels = append(fileHashModels, fileHashModel)
		belogs.Debug("ParseMftModelByOpensslResults(): raw get fileHashModel: ", jsonutil.MarshalJson(fileHashModel))
	}
	mftModel.FileHashModels = fileHashModels
	belogs.Debug("ParseMftModelByOpensslResults():raw get  mftModel:", jsonutil.MarshalJson(mftModel))
	return nil
}

// asID as in rfc6482
type RouteOriginAttestation struct {
	AsID         ASID                 `json:"asID"`
	IpAddrBlocks []ROAIPAddressFamily `json:"ipAddrBlocks"`
}
type ASID int64
type ROAIPAddressFamily struct {
	AddressFamily []byte         `json:"addressFamily"`
	Addresses     []ROAIPAddress `json:"addresses"`
}
type ROAIPAddress struct {
	Address   asn1.BitString `json:"address"`
	MaxLength int            `asn1:"optional,default:-1" json:"maxLength"`
}

type IPAddress asn1.BitString

func ParseRoaModelByOpensslResults(results []string, roaModel *model.RoaModel) (err error) {
	// get roa hex
	// the first HEX DUMP
	/*
				39:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.24
				52:d=4  hl=2 l=inf  cons: cont [ 0 ]
		  	    54:d=5  hl=2 l=inf  cons: OCTET STRING
				56:d=6  hl=2 l=  69 prim: OCTET STRING      [HEX DUMP]:304302026D90303D30270402000130213009030
					BE70400201143009030406BF67000201133009030403C85B20020118301204020002300C300A0305002803EA80020120
	*/
	var roaHex string
	foundAllRoaHex := false
	keyword := "[HEX DUMP]:"
	for i, one := range results {
		if strings.Contains(one, keyword) {
			index := strings.Index(one, keyword)
			roaHex = string(one[index+len(keyword):])
			belogs.Debug("ParseRoaModelByOpensslResults(): len(roaHex):", len(roaHex))

			if !strings.Contains(results[i+1], keyword) {
				foundAllRoaHex = true
				belogs.Debug("ParseRoaModelByOpensslResults(): foundAllRoaHex:", foundAllRoaHex)
				break
			}
			// one [HEX DUMP] length is 10000, so some roa have many ip, will have many [HEX DUMP],
			// just one loop using foundRoaHex to break
			for ii := i + 1; ii < len(results); ii++ {
				if strings.Contains(results[ii], keyword) {
					indexii := strings.Index(results[ii], keyword)
					roaHexii := string(results[ii][indexii+len(keyword):])
					roaHex = roaHex + roaHexii
				} else {
					foundAllRoaHex = true
					break
				}
			}
		}
		if foundAllRoaHex {
			break
		}
	}
	belogs.Debug("ParseRoaModelByOpensslResults():all len(roaHex):", len(roaHex))

	if len(roaHex) == 0 {
		belogs.Error("ParseRoaModelByOpensslResults():len(roaHex) == 0")
		return errors.New("not found roa hex")
	}
	roaByte, err := hex.DecodeString(roaHex)
	if err != nil {
		belogs.Error("ParseRoaModelByOpensslResults():DecodeString err:", err)
		return err
	}
	roa := RouteOriginAttestation{}
	_, err = asn1.Unmarshal(roaByte, &roa)
	if err != nil {
		belogs.Error("ParseRoaModelByOpensslResults():Unmarshal roaByte, err:", err)
		return err
	}
	belogs.Debug("ParseRoaModelByOpensslResults(): roa:", jsonutil.MarshalJson(roa))

	if len(roa.IpAddrBlocks) == 0 {
		belogs.Error("ParseRoaModelByOpensslResults():roa.IpAddrBlocks==0, len(roaByte):", len(roaByte))
		return errors.New("parse roa hex error")
	}

	roaModel.Asn = int64(roa.AsID)
	labRpkiRoaIpaddressParses := make([]model.RoaIpAddressModel, 0)
	for _, one := range roa.IpAddrBlocks {

		for _, ad := range one.Addresses {
			roaIpAddressModel := model.RoaIpAddressModel{}
			roaIpAddressModel.AddressFamily = uint64(one.AddressFamily[1])
			ipAddressTmp := iputil.RoaFormtToIp(ad.Address.Bytes, int(one.AddressFamily[1]))
			if len(ipAddressTmp) == 0 {
				belogs.Error("ParseRoaModelByOpensslResults():RoaFormtToIp ip is empty or too long:",
					"   ad.Address.Bytes ", convert.PrintBytes(ad.Address.Bytes, 8),
					"   address family:", int(one.AddressFamily[1]))
				return errors.New("ad.Address.Bytes is empty or too long")
			}
			prefixLengthTmp := ad.Address.BitLength
			if !iputil.CheckPrefixLengthOrMaxLength(prefixLengthTmp, int(one.AddressFamily[1])) {
				belogs.Error("ParseRoaModelByOpensslResults():CheckPrefixLengthOrMaxLength prefixLength is empty or too long:",
					"   ad.Address.Bytes ", convert.PrintBytes(ad.Address.Bytes, 8),
					"   address family:", int(one.AddressFamily[1]))
				return errors.New("prefixLength is empty or too long")
			}

			roaIpAddressModel.AddressPrefix = ipAddressTmp + "/" + strconv.Itoa(prefixLengthTmp)
			roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, err =
				iputil.AddressPrefixToHexRange(roaIpAddressModel.AddressPrefix, int(roaIpAddressModel.AddressFamily))
			if err != nil {
				belogs.Error("ParseRoaModelByOpensslResults():AddressPrefixToHexRange err:",
					"addressprefix is "+roaIpAddressModel.AddressPrefix, err)
				return err
			}
			addressPrefix, err := iputil.FillAddressPrefixWithZero(roaIpAddressModel.AddressPrefix,
				iputil.GetIpType(roaIpAddressModel.AddressPrefix))
			if err != nil {
				belogs.Error("ParseRoaModelByOpensslResults():FillAddressWithZero err:",
					"addressprefix is "+roaIpAddressModel.AddressPrefix, err)
				return err
			}
			roaIpAddressModel.AddressPrefixRange = jsonutil.MarshalJson(addressPrefix)

			if ad.MaxLength > 0 {
				roaIpAddressModel.MaxLength = uint64(ad.MaxLength)
			} else if ad.MaxLength == 0 {
				// but not return error, will in parsevalidate found and save this error
				belogs.Error("ParseRoaModelByOpensslResults(): MaxLength is zero:",
					jsonutil.MarshalJson(ad))
				roaIpAddressModel.MaxLength = 0
			} else {
				// when ad.Maxlength==-1(default), it is no exist ,then will set as ad.Address.BitLength
				roaIpAddressModel.MaxLength = uint64(ad.Address.BitLength)
			}
			labRpkiRoaIpaddressParses = append(labRpkiRoaIpaddressParses, roaIpAddressModel)
		}
	}
	roaModel.RoaIpAddressModels = labRpkiRoaIpaddressParses
	belogs.Debug("ParseRoaModelByOpensslResults(): roaModel:", jsonutil.MarshalJson(roaModel))

	return nil
}

func ParseMftEContentTypeByOpensslResults(results []string) (eContentType string, err error) {
	// get 1.2.840.113549.1.9.16.1.26
	oid := "1.2.840.113549.1.9.16.1.26"
	for _, one := range results {
		if strings.Contains(one, oid) {
			return oid, nil
		}
	}
	return "", errors.New("invalid content type")
}
func ParseRoaEContentTypeByOpensslResults(results []string) (eContentType string, err error) {
	// get 1.2.840.113549.1.9.16.1.24
	oid := "1.2.840.113549.1.9.16.1.24"
	for _, one := range results {
		if strings.Contains(one, oid) {
			return oid, nil
		}
	}
	return "", errors.New("invalid content type")
}

// parse to get signerInfo
func ParseSignerInfoModelByOpensslResults(results []string) (signerInfoModel model.SignerInfoModel, err error) {
	/*
	   1497:d=5  hl=2 l=   1 prim: INTEGER           :03
	   1500:d=5  hl=2 l=  20 prim: cont [ 0 ]
	   1524:d=6  hl=2 l=   9 prim: OBJECT            :sha256
	   1535:d=5  hl=2 l= 107 cons: cont [ 0 ]
	   1537:d=6  hl=2 l=  26 cons: SEQUENCE
	   1539:d=7  hl=2 l=   9 prim: OBJECT            :contentType
	   1550:d=7  hl=2 l=  13 cons: SET
	   1552:d=8  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.24
	   1565:d=6  hl=2 l=  28 cons: SEQUENCE
	   1567:d=7  hl=2 l=   9 prim: OBJECT            :signingTime
	   1578:d=7  hl=2 l=  15 cons: SET
	   1580:d=8  hl=2 l=  13 prim: UTCTIME           :190601095044Z
	   1595:d=6  hl=2 l=  47 cons: SEQUENCE
	   1597:d=7  hl=2 l=   9 prim: OBJECT            :messageDigest
	   1608:d=7  hl=2 l=  34 cons: SET
	   1610:d=8  hl=2 l=  32 prim: OCTET STRING      [HEX DUMP]:29590CEB666A80B74BAFFD91DC37ADF96BF57D82FCBAF22187FFBED18F898CF
	   1644:d=5  hl=2 l=  13 cons: SEQUENCE
	   1646:d=6  hl=2 l=   9 prim: OBJECT            :rsaEncryption
	   1657:d=6  hl=2 l=   0 prim: NULL
	*/
	signerInfoModel = model.SignerInfoModel{}
	sigStart1 := "d=5  hl=2"
	sigStart2 := "prim: INTEGER"
	sigStartLine := 0

	for i := len(results) - 1; i >= 0; i-- {
		if strings.Contains(results[i], sigStart1) && strings.Contains(results[i], sigStart2) {
			sigStartLine = i

			split := strings.Split(results[i], ":")
			versionStr := split[len(split)-1]
			signerInfoModel.Version, err = strconv.Atoi(strings.TrimSpace(versionStr))
			belogs.Debug("ParseSignerInfoModelByOpensslResults(): signerInfoModel.Version:", signerInfoModel.Version)

			split = strings.Split(results[i+3], ":")
			algStr := split[len(split)-1]
			signerInfoModel.DigestAlgorithm = strings.TrimSpace(algStr)
			belogs.Debug("ParseSignerInfoModelByOpensslResults(): signerInfoModel.DigestAlgorithm:", signerInfoModel.DigestAlgorithm)

			break
		}
	}
	belogs.Debug("ParseSignerInfoModelByOpensslResults(): sigStartLine:", sigStartLine)
	resultsSigs := results[sigStartLine:]
	belogs.Debug("ParseSignerInfoModelByOpensslResults(): len(resultsSigs):", len(resultsSigs))

	sigContentType := "contentType"
	sigTime := "signingTime"
	messageDig := "messageDigest"
	for i := 0; i < len(resultsSigs); i++ {
		if strings.Contains(resultsSigs[i], sigContentType) {
			// next 2 lines, is oid in contentType
			split := strings.Split(resultsSigs[i+2], ":")
			signerInfoModel.ContentType = strings.TrimSpace(split[len(split)-1])
			belogs.Debug("ParseSignerInfoModelByOpensslResults(): signerInfoModel.ContentType:", signerInfoModel.ContentType)
		}

		if strings.Contains(resultsSigs[i], sigTime) {
			// next 2 lines, is oid in signingTime
			split := strings.Split(resultsSigs[i+2], ":")
			sigTimeStr := strings.TrimSpace(split[len(split)-1])
			tm, err := datetime.ParseTime(sigTimeStr, "060102150405Z")
			if err != nil {
				belogs.Error("ParseSignerInfoModelByOpensslResults():datetime.ParseTime err: ", err)
				return signerInfoModel, err
			}
			signerInfoModel.SigningTime = tm
			belogs.Debug("ParseSignerInfoModelByOpensslResults(): signerInfoModel.SigningTime:", signerInfoModel.SigningTime)
		}

		if strings.Contains(resultsSigs[i], messageDig) {
			// next 2 lines, is oid in signingTime
			split := strings.Split(resultsSigs[i+2], ":")
			signerInfoModel.MessageDigest = strings.TrimSpace(split[len(split)-1])
			belogs.Debug("ParseSignerInfoModelByOpensslResults():resultsSigs[i]:", resultsSigs[i],
				"    resultsSigs[i+2]:", resultsSigs[i+2], "    signerInfoModel.MessageDigest:", signerInfoModel.MessageDigest)
		}
	}
	belogs.Debug("ParseSignerInfoModelByOpensslResults(): signerInfoModel:", jsonutil.MarshalJson(signerInfoModel))

	return signerInfoModel, nil
}

func ParseByOpensslAns1ToX509(certFile string, results []string) (cerFile *os.File, fileByte []byte,
	cerStartIndex int, cerEndIndex int, err error) {
	belogs.Debug("ParseByOpensslAns1ToX509(): certFile:", certFile, "   len(results):", len(results))
	// get cer info
	certType := osutil.Ext(certFile)
	/*
	   400:d=4  hl=4 l=1409 cons: SEQUENCE
	   404:d=5  hl=4 l=1129 cons: SEQUENCE
	   408:d=6  hl=2 l=   3 cons: cont [ 0 ]
	   410:d=7  hl=2 l=   1 prim: INTEGER           :02
	   .........
	   1813:d=3  hl=4 l= 428 cons: SET
	   1817:d=4  hl=4 l= 424 cons: SEQUENCE
	   1821:d=5  hl=2 l=   1 prim: INTEGER           :03
	*/
	iStart := 0
	cerStartIndex = 0
	cerEndIndex = 0
	/*
				835579:d=4  hl=5 l=179842 cons: SEQUENCE
				835584:d=5  hl=5 l=179561 cons: SEQUENCE
				835589:d=6  hl=2 l=   3 cons: cont [ 0 ]
				835591:d=7  hl=2 l=   1 prim: INTEGER           :02
						.........
				1015426:d=3  hl=4 l= 428 cons: SET
				1015430:d=4  hl=4 l= 424 cons: SEQUENCE
				1015434:d=5  hl=2 l=   1 prim: INTEGER           :03

		or end
		 1242:d=5  hl=4 l= 257 prim: BIT STRING
		 1503:d=4  hl=2 l=   0 prim: EOC
		 1505:d=3  hl=4 l= 428 cons: SET
		 1509:d=4  hl=4 l= 424 cons: SEQUENCE
		 1513:d=5  hl=2 l=   1 prim: INTEGER           :03
	*/
	cerStartStr1 := "INTEGER           :02"
	cerEndStr1 := "INTEGER           :03"
	cerEndStr2 := "prim: EOC"
	for i := range results {
		belogs.Debug("ParseByOpensslAns1ToX509(): certFile:", certFile, "  i:", i, "   results[i]:", results[i])
		if cerStartIndex == 0 && strings.HasSuffix(results[i], cerStartStr1) {
			one := results[i-3] // last line 3
			end := strings.Index(one, ":")
			cerStartIndex, _ = strconv.Atoi(strings.TrimSpace(string([]byte(one[:end]))))
			iStart = i
			belogs.Debug("ParseByOpensslAns1ToX509():contains cerStartStr1, certFile:", certFile,
				"  i:", i, "   results[i]:", results[i], "   one:", one, "    cerStartIndex:", cerStartIndex)

		} else if cerStartIndex > 0 && i-iStart > 10 && strings.HasSuffix(results[i], cerEndStr1) {
			// (i-iStart) should > 10
			var one string
			if strings.HasSuffix(results[i-3], cerEndStr2) {
				one = results[i-3] // last line 3
			} else {
				one = results[i-2] // last line 2
			}
			end := strings.Index(one, ":")
			cerEndIndex, _ = strconv.Atoi(strings.TrimSpace(string([]byte(one[:end]))))
			belogs.Debug("ParseByOpensslAns1ToX509():contains cerEndStr1, certFile:", certFile,
				"  i:", i, "   results[i]:", results[i], "   one:", one, "    cerEndIndex:", cerEndIndex)
			break
		}
	}
	belogs.Debug("ParseByOpensslAns1ToX509(): certFile:", certFile, "  cerStartIndex:", cerStartIndex, "   cerEndIndex:", cerEndIndex)

	// if not found ,then found again
	if cerStartIndex == 0 || cerEndIndex == 0 {
		cerStartStr1 := "d=4  hl=4"
		cerStartStr2 := "cons: SEQUENCE"
		cerEndStr1 := "d=3  hl=4"
		cerEndStr2 := "cons: SET"

		cerStartIndex = 0
		cerEndIndex = 0
		for _, one := range results {
			if strings.Contains(one, cerStartStr1) && strings.Contains(one, cerStartStr2) {
				end := strings.Index(one, ":")
				cerStartIndex, _ = strconv.Atoi(strings.TrimSpace(string([]byte(one[:end]))))
			} else if cerStartIndex > 0 && strings.Contains(one, cerEndStr1) && strings.Contains(one, cerEndStr2) {
				end := strings.Index(one, ":")
				cerEndIndex, _ = strconv.Atoi(strings.TrimSpace(string([]byte(one[:end]))))
				break
			}
		}
		belogs.Debug("ParseByOpensslAns1ToX509():again  certFile:", certFile, "  cerStartIndex:", cerStartIndex, "   cerEndIndex:", cerEndIndex)
	}

	if cerStartIndex == 0 || cerEndIndex == 0 {
		belogs.Error("ParseByOpensslAns1ToX509():cerStartIndex==0 || cerEndIndex == 0 , certFile:", certFile)
		return nil, nil, 0, 0, errors.New("fail to parse ee certificate")
	}

	_, fileDecodeBase64Byte, err := asn1util.ReadFileAndDecodeBase64(certFile)
	if err != nil {
		belogs.Error("ParseByOpensslAns1ToX509():ReadFile return err: ", certFile, err)
		return nil, nil, 0, 0, err
	}
	fileByte = fileDecodeBase64Byte[cerStartIndex:cerEndIndex]
	belogs.Debug("ParseByOpensslAns1ToX509():len(fileByte):", len(fileByte))

	// will test file(no trim)
	cerFile, err = ioutil.TempFile("", certType+"_notrim_") // temp file
	if err != nil {
		belogs.Error("ParseByOpensslAns1ToX509():ioutil.TempFile notrim cerFile fail: ", certFile, cerFile, err)
		return nil, nil, cerStartIndex, cerEndIndex, err
	}
	belogs.Debug("ParseByOpensslAns1ToX509():notrim cerFile: [cerStartIndex:cerEndIndex]:", certFile, cerFile.Name(), cerStartIndex, cerEndIndex)
	cerFile.Write(fileByte)
	// test notrim by openssl x509
	_, err = opensslutil.GetResultsByOpensslX509(cerFile.Name())
	if err != nil {
		belogs.Debug("ParseByOpensslAns1ToX509():notrim cerFile fail: [cerStartIndex:cerEndIndex]:",
			certFile, cerFile.Name(),
			"  cerStartIndex, cerEndIndex:", cerStartIndex, cerEndIndex, err)
		osutil.CloseAndRemoveFile(cerFile)

		// test if need trim00
		fileByteTrim, cerEndIndexTrim := asn1util.TrimSuffix00(fileByte, cerEndIndex)
		belogs.Debug("ParseByOpensslAns1ToX509():TrimSuffix00 len(fileByteTrim):", certFile, len(fileByteTrim))

		cerFile, err = ioutil.TempFile("", certType+"_trim_") // temp file
		if err != nil {
			belogs.Error("ParseByOpensslAns1ToX509():ioutil.TempFile trim cerFile: ", certFile, cerFile, err)
			return nil, nil, 0, 0, err
		}
		belogs.Debug("ParseByOpensslAns1ToX509():trim cerFile: [cerStartIndex:cerEndIndexTrim]:", certFile, cerFile.Name(), cerStartIndex, cerEndIndexTrim)
		cerFile.Write(fileByteTrim)

		// test if need trim
		_, err = opensslutil.GetResultsByOpensslX509(cerFile.Name())
		if err != nil {
			// if trim fil, the remove old file(trim)
			belogs.Error("ParseByOpensslAns1ToX509():GetResultsByOpensslX509 trim and notrim all fail: ",
				certFile, cerFile.Name(),
				"   cerStartIndex, cerEndIndex:", cerStartIndex, cerEndIndex,
				"   cerStartIndex, cerEndIndexTrim:", cerStartIndex, cerEndIndexTrim, err)
			osutil.CloseAndRemoveFile(cerFile)
			return nil, nil, 0, 0, err
		}

		belogs.Debug("ParseByOpensslAns1ToX509():trim cerFile pass: [cerStartIndex:cerEndIndexTrim]:", certFile, cerFile.Name(), cerStartIndex, cerEndIndexTrim)
		return cerFile, fileByteTrim, cerStartIndex, cerEndIndexTrim, nil

	}
	belogs.Debug("ParseByOpensslAns1ToX509():notrim cerFile pass: [cerStartIndex:cerEndIndex]:", certFile, cerFile.Name(), cerStartIndex, cerEndIndex)
	return cerFile, fileByte, cerStartIndex, cerEndIndex, nil

}

func ParseCrlModelByOpensslResults(results []string, crlModel *model.CrlModel) (err error) {
	//
	/*
			453:d=5  hl=2 l=   3 prim: OBJECT            :X509v3 Authority Key Identifier
		    458:d=5  hl=2 l=  24 prim: OCTET STRING      [HEX DUMP]:301680148278F47DEC5B7ADC201897F99BCC6E2BFA88D015
	*/
	// AKI crlNum
	akiFound := false
	crlNumFound := false
	for i, one := range results {
		if strings.Contains(one, ":X509v3 Authority Key Identifier") {
			split := strings.Split(results[i+1], ":")
			tmp := strings.TrimSpace(strings.ToLower(split[len(split)-1]))
			crlModel.Aki = string([]byte(tmp)[8:]) // fix 8 length as asn.1 sequence
			akiFound = true
		}
		if strings.Contains(one, ":X509v3 CRL Number") {
			tmp := results[i+1]
			index := strings.Index(tmp, "[HEX DUMP]:")
			tmp = string([]byte(tmp)[index+len("[HEX DUMP]:")+4:]) // fix 8 length as asn.1 sequence
			crlModel.CrlNumber, _ = strconv.ParseUint(tmp, 16, 0)
			crlNumFound = true
		}
		if strings.Contains(one, ":sha256WithRSAEncryption") {
			if crlNumFound {
				crlModel.TbsAlgorithm = "sha256WithRSAEncryption"
			} else {
				crlModel.CertAlgorithm = "sha256WithRSAEncryption"
			}
		}
	}
	if !akiFound {
		belogs.Error("ParseCrlModelByOpensslResults():not found aki: ", results)
		return errors.New("not found aki")
	}
	if !crlNumFound {
		belogs.Error("ParseCrlModelByOpensslResults():not found crl number: ", results)
		return errors.New("not found crl number")
	}
	return nil
}

func ParseSigModelByOpensslResults(results []string, sigModel *model.SigModel) (err error) {
	// openssl asn1parse  -in checklist.sig --inform der
	// the first HEX DUMP
	/*
	    0:d=0  hl=4 l=1703 cons: SEQUENCE
	    4:d=1  hl=2 l=   9 prim: OBJECT            :pkcs7-signedData
	   15:d=1  hl=4 l=1688 cons: cont [ 0 ]
	   19:d=2  hl=4 l=1684 cons: SEQUENCE
	   23:d=3  hl=2 l=   1 prim: INTEGER           :03
	   26:d=3  hl=2 l=  13 cons: SET
	   28:d=4  hl=2 l=  11 cons: SEQUENCE
	   30:d=5  hl=2 l=   9 prim: OBJECT            :sha256
	   41:d=3  hl=3 l= 176 cons: SEQUENCE
	   44:d=4  hl=2 l=   9 prim: OBJECT            :1.3.6.1.4.1.41948.49
	   55:d=4  hl=3 l= 162 cons: cont [ 0 ]
	   58:d=5  hl=3 l= 159 prim: OCTET STRING      [HEX DUMP]:30819C3014A1123010300E04010230090307002001067C208C300B06096086480165030402013077303416106234325F697076365F6C6F612E706E6704209516DD64BE7C1725B9FCA117120E58E8D842A5206873399B3DDFFC91C4B6ACF0303F161B6234325F736572766963655F646566696E6974696F6E2E6A736F6E04200AE1394722005CD92F4C6AA024D5D6B3E2E67D629F11720D9478A633A117A1C7	*/
	var sigHex string
	foundAllSigHex := false
	keyword := "[HEX DUMP]:"
	for i, one := range results {
		if strings.Contains(one, keyword) {
			index := strings.Index(one, keyword)
			sigHex = string(one[index+len(keyword):])
			belogs.Debug("ParseSigModelByOpensslResults(): len(sigHex):", len(sigHex))

			if !strings.Contains(results[i+1], keyword) {
				foundAllSigHex = true
				belogs.Debug("ParseSigModelByOpensslResults(): foundAllSigHex:", foundAllSigHex)
				break
			}
			// one [HEX DUMP] length is more than 10000,  will have many [HEX DUMP],
			// just one loop to break
			for ii := i + 1; ii < len(results); ii++ {
				if strings.Contains(results[ii], keyword) {
					indexii := strings.Index(results[ii], keyword)
					sigHexii := string(results[ii][indexii+len(keyword):])
					sigHex = sigHex + sigHexii
				} else {
					foundAllSigHex = true
					break
				}
			}
		}
		if foundAllSigHex {
			break
		}
	}
	belogs.Debug("ParseSigModelByOpensslResults():all len(sigHex):", len(sigHex), sigHex)
	belogs.Info("ParseSigModelByOpensslResults():len(sigHex):", len(sigHex))

	if len(sigHex) == 0 {
		belogs.Error("ParseSigModelByOpensslResults():len(sigHex) == 0")
		return errors.New("not found sig hex")
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		belogs.Error("ParseSigModelByOpensslResults():DecodeString err:", err)
		return err
	}
	node, err := asn1node.ParseBytes(sigBytes)
	if err != nil {
		belogs.Error("ParseSigModelByOpensslResults():ParseBytes err, sigHex:", sigHex, err)
		return err
	}
	if node == nil {
		belogs.Error("ParseRoaModelByOpensslResults():ParseBytes node==nil,  sigHex:", sigHex)
		return errors.New("parse sig hex error")
	}
	belogs.Debug("ParseSigModelByOpensslResults():node:", jsonutil.MarshalJson(node))

	// get address
	if len(node.Nodes) > 0 && len(node.Nodes[0].Nodes) > 0 && len(node.Nodes[0].Nodes[0].Data) > 0 {
		belogs.Debug("ParseSigModelByOpensslResults():get address :", convert.PrintBytesOneLine(node.Nodes[0].Nodes[0].Data))

		data := node.Nodes[0].Nodes[0].Data
		ipAddrBlocks, err := asn1cert.ParseToIpAddressBlocks(data)
		if err != nil {
			belogs.Error("ParseSigModelByOpensslResults():ParseToIpAddressBlocks err :", convert.PrintBytesOneLine(data), err)
			return err
		}
		belogs.Debug("ParseSigModelByOpensslResults(): ipAddrBlocks:", jsonutil.MarshalJson(ipAddrBlocks))
		sigModel.RpkiSignedChecklist.CerIpAddresses = make([]model.CerIpAddress, 0)
		for i := range ipAddrBlocks {
			cerIpAddress := model.CerIpAddress{
				AddressFamily: ipAddrBlocks[i].AddressFamily,
				AddressPrefix: ipAddrBlocks[i].AddressPrefix,
				Min:           ipAddrBlocks[i].Min,
				Max:           ipAddrBlocks[i].Max,
			}
			sigModel.RpkiSignedChecklist.CerIpAddresses = append(sigModel.RpkiSignedChecklist.CerIpAddresses,
				cerIpAddress)
			belogs.Debug("ParseSigModelByOpensslResults(): cerIpAddress:", jsonutil.MarshalJson(cerIpAddress))
		}
	}
	// get oid
	if len(node.Nodes) > 1 && len(node.Nodes[1].Nodes) > 0 {
		belogs.Debug("ParseSigModelByOpensslResults():get oid :", convert.PrintBytesOneLine(node.Nodes[1].Nodes[0].Data))

		data := node.Nodes[1].Nodes[0].Data
		oid, err := asn1base.ParseObjectIdentifier(data)
		if err != nil {
			belogs.Error("ParseSigModelByOpensslResults():ParseObjectIdentifier err :", convert.PrintBytesOneLine(data), err)
			return err
		}

		sigModel.RpkiSignedChecklist.DigestAlgorithmIdentifier = oid.String()
		belogs.Debug("ParseSigModelByOpensslResults(): DigestAlgorithmIdentifier:", oid.String())
	}

	// get filehash
	if len(node.Nodes) > 2 {
		belogs.Debug("ParseSigModelByOpensslResults():get filehash :", convert.PrintBytesOneLine(node.Nodes[2].FullData))

		data := node.Nodes[2].FullData
		fileAndHashs, err := asn1cert.ParseToFileAndHashs(data)
		if err != nil {
			belogs.Error("ParseSigModelByOpensslResults():ParseToFileAndHashs err :", convert.PrintBytesOneLine(data), err)
			return err
		}
		sigModel.RpkiSignedChecklist.FileHashModels = make([]model.FileHashModel, 0)
		for i := range fileAndHashs {
			fh := model.FileHashModel{
				File: fileAndHashs[i].File,
				Hash: strings.Replace(convert.PrintBytesOneLine(fileAndHashs[i].Hash), " ", "", -1),
			}
			sigModel.RpkiSignedChecklist.FileHashModels = append(sigModel.RpkiSignedChecklist.FileHashModels, fh)
			belogs.Debug("ParseSigModelByOpensslResults(): FileHashModel:", jsonutil.MarshalJson(fh))
		}
	}

	return nil
}
func ParseSigEContentTypeByOpensslResults(results []string) (eContentType string, err error) {
	// get 1.3.6.1.4.1.41948.49
	/*
		1329:d=7  hl=2 l=   9 prim: OBJECT            :contentType
		1340:d=7  hl=2 l=  11 cons: SET
		1342:d=8  hl=2 l=   9 prim: OBJECT            :1.3.6.1.4.1.41948.49
	*/
	oid := "1.3.6.1.4.1.41948.49"
	for _, one := range results {
		if strings.Contains(one, oid) {
			return oid, nil
		}
	}
	return "", errors.New("invalid content type")
}

func ParseAsaModelByOpensslResults(results []string, asaModel *model.AsaModel) (err error) {
	// openssl asn1parse  -in checklist.asa --inform der
	// the first HEX DUMP
	/*
			openssl asn1parse -in AS211321.asa  -inform DER
			    0:d=0  hl=4 l=1734 cons: SEQUENCE
			    4:d=1  hl=2 l=   9 prim: OBJECT            :pkcs7-signedData
			    15:d=1  hl=4 l=1719 cons: cont [ 0 ]
			    19:d=2  hl=4 l=1715 cons: SEQUENCE
			    23:d=3  hl=2 l=   1 prim: INTEGER           :03
			    26:d=3  hl=2 l=  13 cons: SET
			    28:d=4  hl=2 l=  11 cons: SEQUENCE
			    30:d=5  hl=2 l=   9 prim: OBJECT            :sha256
			    41:d=3  hl=2 l=  55 cons: SEQUENCE
		 	    43:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.49
				56:d=4  hl=2 l=  40 cons: cont [ 0 ]
				58:d=5  hl=2 l=  38 prim: OCTET STRING      [HEX DUMP]:30240203033979301D3005020300FDE83009020300FDE9040200013009020300FDEA04020002
	*/
	var asaHex string
	foundAllAsaHex := false
	keyword := "[HEX DUMP]:"
	for i, one := range results {
		if strings.Contains(one, keyword) {
			index := strings.Index(one, keyword)
			asaHex = string(one[index+len(keyword):])
			belogs.Debug("ParseAsaModelByOpensslResults(): len(asaHex):", len(asaHex))

			if !strings.Contains(results[i+1], keyword) {
				foundAllAsaHex = true
				belogs.Debug("ParseAsaModelByOpensslResults(): foundAllAsaHex:", foundAllAsaHex)
				break
			}
			// one [HEX DUMP] length is more than 10000,  will have many [HEX DUMP],
			// just one loop to break
			for ii := i + 1; ii < len(results); ii++ {
				if strings.Contains(results[ii], keyword) {
					indexii := strings.Index(results[ii], keyword)
					asaHexii := string(results[ii][indexii+len(keyword):])
					asaHex = asaHex + asaHexii
				} else {
					foundAllAsaHex = true
					break
				}
			}
		}
		if foundAllAsaHex {
			break
		}
	}
	belogs.Debug("ParseAsaModelByOpensslResults():all len(asaHex):", len(asaHex), asaHex)
	belogs.Info("ParseAsaModelByOpensslResults():len(asaHex):", len(asaHex))

	if len(asaHex) == 0 {
		belogs.Error("ParseAsaModelByOpensslResults():len(asaHex) == 0")
		return errors.New("not found asa hex")
	}
	asaBytes, err := hex.DecodeString(asaHex)
	if err != nil {
		belogs.Error("ParseAsaModelByOpensslResults():DecodeString err:", err)
		return err
	}
	asProviderAttestation := AsProviderAttestation{}
	_, err = asn1.Unmarshal(asaBytes, &asProviderAttestation)
	if err != nil {
		belogs.Error("ParseAsaModelByOpensslResults():asn1.Unmarshal err:", err)
		return err
	}
	asaModel.CustomerAsns, err = convertAsProviderAttestationToCustomerAsns(asProviderAttestation)
	if err != nil {
		belogs.Error("ParseAsaModelByOpensslResults():convertAsProviderAttestationToCustomerAsns err:", err)
		return err
	}
	belogs.Info("ParseAsaModelByOpensslResults(): asaModel.CustomerAsns:", jsonutil.MarshalJson(asaModel.CustomerAsns))
	return nil
}

func ParseAsaEContentTypeByOpensslResults(results []string) (eContentType string, err error) {
	// get 1.2.840.113549.1.9.16.1.49
	/*
		1358:d=7  hl=2 l=   9 prim: OBJECT            :contentType
		1369:d=7  hl=2 l=  13 cons: SET
		1371:d=8  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.49
	*/
	oid := "1.2.840.113549.1.9.16.1.49"
	for _, one := range results {
		if strings.Contains(one, oid) {
			return oid, nil
		}
	}
	return "", errors.New("invalid content type")
}
