package parsevalidatecore

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	asn1 "github.com/bgpsecurity/rpstir2/parsevalidate-asn1"
	openssl "github.com/bgpsecurity/rpstir2/parsevalidate-openssl"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func parseValidateRoa(fileModel *model.FileModel) (roaModel model.RoaModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateRoa(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseRoaModel(fileModel, &roaModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateRoa(): parseRoaModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return roaModel, stateModel, err
	}
	belogs.Debug("parseValidateRoa(): roaModel:", roaModel.String(), "  time(s):", time.Since(start))

	err = validateRoaModel(&roaModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateRoa():validateRoaModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return roaModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateRoa():stateModel have errors or warnings, fileModel:", jsonutil.MarshalJson(fileModel), "     stateModel:", jsonutil.MarshalJson(stateModel))
	}
	stateModel.JudgeState()
	belogs.Debug("parseValidateRoa():  roaModel.FilePath, roaModel.FileName, roaModel.Ski, roaModel.Aki:",
		roaModel.FilePath, roaModel.FileName, roaModel.Ski, roaModel.Aki, "  time(s):", time.Since(start))
	return roaModel, stateModel, nil
}

func parseRoaModel(fileModel *model.FileModel, roaModel *model.RoaModel, stateModel *model.StateModel) error {
	belogs.Debug("parseRoaModel(): fileModel:", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseRoaModel(): GetFileLength fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseRoaModel(): fileLength:", fileLength, " tempFile:", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseRoaModel(): GetFileLength fail, fileLenght is empty, fileModel:",
		 jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	roaModel.FilePath, roaModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	//get file hash
	roaModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseRoaModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	err = asn1.ParseRoaModelByAsn1(fileModel, roaModel, stateModel)
	if err != nil {
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("parseRoaModel(): ParseRoaModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("parseRoaModel(): ParseRoaModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		} // have set stateMsg in ParseRoaModelByAsn1 .
		err = openssl.ParseRoaModelByOpenssl(fileModel, roaModel, stateModel)
		if err != nil {
			belogs.Error("parseRoaModel(): ParseRoaModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			// have set stateMsg in ParseRoaModelByOpenssl .
			return err
		}
	}

	/*
		//https://blog.csdn.net/Zhymax/article/details/7683925
		// get asn1 using to cer、crl
		//openssl asn1parse -in -0AU6cJZAl7QHJeNhN9vE3zUBr4.roa -inform DER
		results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseRoaModel(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseRoaModel(): len(results):", len(results))

		err = openssl.ParseRoaModelByOpensslResults(results, roaModel)
		if err != nil {
			belogs.Error("parseRoaModel(): ParseRoaModelByOpensslResults fail, will try parseRoaModelByPacket, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)

			err = parseRoaModelByPacket(fileModel, roaModel)
			if err != nil {
				belogs.Error("parseRoaModel(): parseRoaModelByPacket fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				return err
			}
		}

		roaModel.EContentType, err = openssl.ParseRoaEContentTypeByOpensslResults(results)
		if err != nil {
			belogs.Error("parseRoaModel():ParseEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		roaModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseRoaModel():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// get cer info in roa
		eeCerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
		if err != nil {
			belogs.Error("parseRoaModel():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		defer osutil.CloseAndRemoveFile(eeCerFile)

		results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
		if err != nil {
			belogs.Error("parseRoaModel(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseRoaModel(): len(results):", len(results))

		roaModel.Aki, roaModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
		if err != nil {
			belogs.Error("parseRoaModel(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// AIA SIA
		roaModel.AiaModel, roaModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseRoaModel(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// EE
		roaModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, start, end)
		if err != nil {
			belogs.Error("parseRoaModel(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse roa to get ee",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// get IP address in EE
		roaModel.EeCertModel.CerIpAddressModel, _, err = openssl.ParseCerIpAddressModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseRoaModel(): ParseCerIpAddressModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
	*/
	belogs.Debug("parseRoaModel(): roaModel:", roaModel.String())
	return nil
}

/*
func parseRoaModelByPacket(fileModel *model.FileModel, roaModel *model.RoaModel) error {

	fileByte, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseRoaModelByPacket():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return err
	}
	//get file hash
	roaModel.FileHash = hashutil.Sha256(fileByte)

	pack := packet.DecodePacket(fileDecodeBase64Byte)
	//packet.PrintPacketString("parseRoaModelByPacket():DecodePacket: ", pack, true, true)

	var oidPacketss = &[]packet.OidPacket{}
	packet.TransformPacket(pack, oidPacketss)
	packet.PrintOidPacket(oidPacketss)

	err = packet.ExtractRoaOid(oidPacketss, fileModel.TempFilePathName, fileDecodeBase64Byte, roaModel)
	if err != nil {
		belogs.Error("parseRoaModelByPacket():ExtractRoaOid fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return err
	}

	return nil
}
*/

// only validate roa self.  in chain check, will check fathers;;;;
// https://datatracker.ietf.org/doc/rfc6482/?include_text=1  A Profile for Route Origin Authorizations (ROAs)
// https://datatracker.ietf.org/doc/rfc6488/?include_text=1
// shqhl.c P1955 verify_roa() -->  roa_validate.c  roaValidate() and roaValidate2() ,
// TODO but roaValidate2 is too strange, not understand yet
func validateRoaModel(roaModel *model.RoaModel, stateModel *model.StateModel) (err error) {

	if roaModel.Version != 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Wrong Version number",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	if roaModel.Asn < 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ASN is negative",
			Detail: fmt.Sprintf("Asn is %d", roaModel.Asn)}
		stateModel.AddError(&stateMsg)
	}
	if roaModel.Asn > 0xffffffff {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ASN is too large",
			Detail: fmt.Sprintf("Asn is %d", roaModel.Asn)}
		stateModel.AddError(&stateMsg)
	}

	if len(roaModel.RoaIpAddressModels) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "There is no IP address",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// ski aki
	if len(roaModel.Ski) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(roaModel.Ski) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(roaModel.Aki) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(roaModel.Aki) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// check min, max
	// TODO roa_vlidate.c P396 setup_roa_minmax	P344 setup_cert_minmax
	for _, one := range roaModel.RoaIpAddressModels {
		if one.AddressFamily != 1 && one.AddressFamily != 2 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "IP address is neither IPv4 nor IPv6",
				Detail: fmt.Sprintf("family is %d", one.AddressFamily)}
			stateModel.AddError(&stateMsg)
		}
		if !iputil.IsAddressPrefix(one.AddressPrefix) {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "IP address format is wrong",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}

		_, prefix, err := iputil.SplitAddressAndPrefix(one.AddressPrefix)
		if err != nil {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "IP address format is wrong",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if one.MaxLength == 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Maxlength of IP address is zero",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if one.MaxLength != 0 && one.MaxLength < prefix {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail: "Maxlength of IP address is smaller than prefix length",
				Detail: "maxlength is " + convert.ToString(one.MaxLength) +
					", and prefix is " + convert.ToString(prefix)}
			stateModel.AddError(&stateMsg)
		}

		if one.AddressFamily == 1 {
			if len(one.AddressPrefix) > 18 { // 255.255.255.255/32
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "IPv4 address format is wrong",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
			if one.MaxLength > 32 {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Maxlength of IPv4 address is too large",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
		}
		if one.AddressFamily == 2 {
			if len(one.AddressPrefix) > 49 { //ffff:ffff:ffff:ffff:ffff:ffff:255:255:255:255/128
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "IPv6 address format is wrong",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
			if one.MaxLength > 128 {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Maxlength of IPv6 address is too large",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
		}

		// check in ee cert ipAddress ,have same addressprefix( no maxlength)
		found := false
		for _, cip := range roaModel.EeCertModel.CerIpAddressModel.CerIpAddresses {
			// compare directly
			if one.AddressPrefix == cip.AddressPrefix {
				found = true
				break
			}
			// compare range
			if one.RangeStart == cip.RangeStart && one.RangeEnd == cip.RangeEnd {
				found = true
				break
			}
			// ip in ee is larger than ip in roa
			// cip.RangeStart <--- one.RangeStart <---------> one.RangeEnd ---> cip.RangeEnd
			if cip.RangeStart <= one.RangeStart && one.RangeEnd <= cip.RangeEnd {
				found = true
				break
			}

			// trim zero in ee ip, then compare
			cipTrim, err := iputil.TrimAddressPrefixZero(cip.AddressPrefix, int(cip.AddressFamily))
			if err != nil {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "IP address of EE format is wrong",
					Detail: "" + cip.AddressPrefix}
				stateModel.AddError(&stateMsg)
			}
			if one.AddressPrefix == cipTrim {
				found = true
				break
			}

		}
		if !found {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "IP address is not in IP address of EE range",
				Detail: "roa ip address is " + one.AddressPrefix + "[" + one.RangeStart + ":" + one.RangeEnd + "]"}
			stateModel.AddError(&stateMsg)
		}
	}

	// check time
	ValidateEeCertModel(stateModel, &roaModel.EeCertModel)
	ValidateSignerInfoModel(stateModel, &roaModel.SignerInfoModel)

	belogs.Debug("validateRoaModel():filePath, fileName,stateModel:",
		roaModel.FilePath, roaModel.FileName, jsonutil.MarshalJson(stateModel))

	return nil
}
