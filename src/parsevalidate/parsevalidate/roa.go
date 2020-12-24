package parsevalidate

import (
	"fmt"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	hashutil "github.com/cpusoft/goutil/hashutil"
	iputil "github.com/cpusoft/goutil/iputil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	opensslutil "github.com/cpusoft/goutil/opensslutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"model"
	"parsevalidate/openssl"
	"parsevalidate/packet"
	"parsevalidate/util"
)

//Try to store the error in statemode instead of returning err
func ParseValidateRoa(certFile string) (roaModel model.RoaModel, stateModel model.StateModel, err error) {
	stateModel = model.NewStateModel()
	err = parseRoaModel(certFile, &roaModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateRoa(): parseRoaByOpenssl err:", certFile, err)
		return roaModel, stateModel, nil
	}

	err = validateRoaModel(&roaModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateRoa():validateRoaModel err:", certFile, err)
		return roaModel, stateModel, nil
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("ParseValidateRoa():stateModel have errors or warnings", certFile, "     stateModel:", jsonutil.MarshalJson(stateModel))
	}

	belogs.Debug("ParseValidateRoa():  roaModel.FilePath, roaModel.FileName, roaModel.Ski, roaModel.Aki:",
		roaModel.FilePath, roaModel.FileName, roaModel.Ski, roaModel.Aki)
	return roaModel, stateModel, nil
}

func parseRoaModel(certFile string, roaModel *model.RoaModel, stateModel *model.StateModel) error {
	belogs.Debug("parseRoaModel(): certFile: ", certFile)
	roaModel.FilePath, roaModel.FileName = osutil.GetFilePathAndFileName(certFile)

	//https://blog.csdn.net/Zhymax/article/details/7683925
	// get asn1 using to cer、crl
	//openssl asn1parse -in -0AU6cJZAl7QHJeNhN9vE3zUBr4.roa -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(certFile)
	if err != nil {
		belogs.Error("parseRoaModel(): GetResultsByOpensslAns1: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseRoaModel(): len(results):", len(results))

	//get file hash
	roaModel.FileHash, err = hashutil.Sha256File(certFile)
	if err != nil {
		belogs.Error("parseRoaModel(): Sha256File: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	err = openssl.ParseRoaModelByOpensslResults(results, roaModel)
	if err != nil {
		belogs.Error("parseRoaModel(): ParseRoaModelByOpensslResults:  certFile:", certFile, "  err:", err, " will try parseMftModelByPacket")
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)

		err = parseRoaModelByPacket(certFile, roaModel)
		if err != nil {
			belogs.Error("parseRoaModel(): parseRoaModelByPacket err:", certFile, err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
	}

	roaModel.EContentType, err = openssl.ParseRoaEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("parseRoaModel():ParseEContentTypeByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	roaModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseRoaModel():ParseSignerInfoModelByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in roa
	cerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(certFile, results)
	if err != nil {
		belogs.Error("parseRoaModel():ParseByOpensslAns1ToX509  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(cerFile)

	results, err = opensslutil.GetResultsByOpensslX509(cerFile.Name())
	if err != nil {
		belogs.Error("parseRoaModel(): GetResultsByOpensslX509: err: ", err, ": "+cerFile.Name())
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseRoaModel(): len(results):", len(results))

	roaModel.Aki, roaModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseRoaModel(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA SIA
	roaModel.AiaModel, roaModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseRoaModel(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// EE
	roaModel.EeCertModel, err = ParseEeCertModel(cerFile.Name(), fileByte, start, end)
	if err != nil {
		belogs.Error("parseRoaModel(): ParseEeCertModel: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "parse roa to get ee fail",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get IP address in EE
	roaModel.EeCertModel.CerIpAddressModel, _, err = openssl.ParseCerIpAddressModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseRoaModel(): ParseCerIpAddressModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	belogs.Debug("parseRoaModel(): roaModel:", jsonutil.MarshalJson(roaModel))
	return nil
}

func parseRoaModelByPacket(certFile string, roaModel *model.RoaModel) error {

	fileByte, fileDecodeBase64Byte, err := util.ReadFileAndDecodeBase64(certFile)
	if err != nil {
		belogs.Error("parseRoaModelByPacket():ReadFile return err: ", certFile, err)
		return err
	}
	//get file hash
	roaModel.FileHash = hashutil.Sha256(fileByte)

	pack := packet.DecodePacket(fileDecodeBase64Byte)
	//packet.PrintPacketString("parseRoaModelByPacket():DecodePacket: ", pack, true, true)

	var oidPacketss = &[]packet.OidPacket{}
	packet.TransformPacket(pack, oidPacketss)
	packet.PrintOidPacket(oidPacketss)

	err = packet.ExtractRoaOid(oidPacketss, certFile, fileDecodeBase64Byte, roaModel)
	if err != nil {
		belogs.Error("parseRoaModelByPacket():ExtractRoaOid err:", err)
		return err
	}

	return nil
}

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
			Detail: fmt.Sprint("Asn is %d", roaModel.Asn)}
		stateModel.AddError(&stateMsg)
	}
	if roaModel.Asn > 0xffffffff {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ASN is too large",
			Detail: fmt.Sprint("Asn is %d", roaModel.Asn)}
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
				Detail: fmt.Sprint("family is %d", one.AddressFamily)}
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
