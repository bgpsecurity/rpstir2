package openssl

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	packet "github.com/bgpsecurity/rpstir2/parsevalidate-packet"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
)

func ParseRoaModelByOpenssl(fileModel *model.FileModel, roaModel *model.RoaModel, stateModel *model.StateModel) error {
	start := time.Now()
	belogs.Debug("ParseRoaModelByOpenssl(): fileModel:", jsonutil.MarshalJson(fileModel))
	//https://blog.csdn.net/Zhymax/article/details/7683925
	// get asn1 using to cer、crl
	//openssl asn1parse -in -0AU6cJZAl7QHJeNhN9vE3zUBr4.roa -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseRoaModelByOpenssl(): len(results):", len(results))

	err = ParseRoaModelByOpensslResults(results, roaModel)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): ParseRoaModelByOpensslResults fail, will try parseRoaModelByPacket, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)

		err = parseRoaModelByPacket(fileModel, roaModel)
		if err != nil {
			belogs.Error("ParseRoaModelByOpenssl(): parseRoaModelByPacket fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
	}

	roaModel.EContentType, err = ParseRoaEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl():ParseEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	roaModel.SignerInfoModel, err = ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in roa
	eeCerFile, fileByte, eeCertStart, eeCertEnd, err := ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(eeCerFile)

	results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseRoaModelByOpenssl(): len(results):", len(results))

	roaModel.Aki, roaModel.Ski, err = ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA SIA
	roaModel.AiaModel, roaModel.SiaModel, err = ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// EE
	roaModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, eeCertStart, eeCertEnd)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse roa to get ee",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get IP address in EE
	roaModel.EeCertModel.CerIpAddressModel, _, err = ParseCerIpAddressModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseRoaModelByOpenssl(): ParseCerIpAddressModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	belogs.Debug("ParseRoaModelByOpenssl(): roaModel:", roaModel.String(), "  time(s):", time.Since(start))
	return nil
}

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
