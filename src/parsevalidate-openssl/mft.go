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

func ParseMftModelByOpenssl(fileModel *model.FileModel, mftModel *model.MftModel, stateModel *model.StateModel) error {
	start := time.Now()
	belogs.Debug("ParseMftModelByOpenssl(): fileModel:", jsonutil.MarshalJson(fileModel))
	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.mft -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseMftModelByOpenssl(): len(results):", len(results))

	// get mft hex
	// first HEX DUMP
	/*
		   39:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.26
		   52:d=4  hl=2 l=inf  cons: cont [ 0 ]
		   54:d=5  hl=2 l=inf  cons: OCTET STRING
		   56:d=6  hl=3 l= 137 prim: OCTET STRING      [HEX DUMP]:308186020200CA180F323031383036323831373030
		32345A180F32303138303632393138303032345A060960864801650304020130533051162C36353736393433633735383262
		3164656266666261303564363235343034323462633765626363352E63726C032100154269177B0346014642A367DA415F32
		C2BFE7C4EAD8AED59ACCF8F20220F89C
	*/

	err = ParseMftModelByOpensslResults(results, mftModel)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl():ParseMftModelByOpensslResults fail, will try parseMftModelByPacket, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)

		err = parseMftModelByPacket(fileModel, mftModel)
		if err != nil {
			belogs.Error("ParseMftModelByOpenssl():parseMftModelByPacket fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

	}

	mftModel.EContentType, err = ParseMftEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl():ParseEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	mftModel.SignerInfoModel, err = ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in mft
	eeCerFile, fileByte, eeCertStart, eeCertEnd, err := ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(eeCerFile)

	results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseMftModelByOpenssl(): len(results):", len(results))

	mftModel.Aki, mftModel.Ski, err = ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA SIA
	mftModel.AiaModel, mftModel.SiaModel, err = ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// EE
	mftModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, eeCertStart, eeCertEnd)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	// get IP address in EE: RFC9286, will check ipaddress(should be empty)
	mftModel.EeCertModel.CerIpAddressModel, _, err = ParseCerIpAddressModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseMftModelByOpenssl(): ParseCerIpAddressModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	belogs.Debug("ParseMftModelByOpenssl(): mftModel:", mftModel.String(), "  time(s):", time.Since(start))
	return nil
}
func parseMftModelByPacket(fileModel *model.FileModel, mftModel *model.MftModel) error {

	fileByte, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseMftModelByPacket():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return err
	}

	//get file hash
	mftModel.FileHash = hashutil.Sha256(fileByte)

	pack := packet.DecodePacket(fileDecodeBase64Byte)
	//packet.PrintPacketString("parseMftModelByPacket():DecodePacket: ", pack, true, true)

	var oidPacketss = &[]packet.OidPacket{}
	packet.TransformPacket(pack, oidPacketss)
	packet.PrintOidPacket(oidPacketss)

	// manifests,
	err = packet.ExtractMftOid(oidPacketss, fileModel.TempFilePathName, fileDecodeBase64Byte, mftModel)
	if err != nil {
		belogs.Error("parseMftModelByPacket():ExtractMftOid fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return err
	}

	return nil
}
