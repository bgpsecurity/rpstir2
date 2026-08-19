package openssl

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
)

func ParseSigModelByOpenssl(fileModel *model.FileModel, sigModel *model.SigModel, stateModel *model.StateModel) error {
	start := time.Now()
	belogs.Debug("ParseSigModelByOpenssl(): fileModel:", jsonutil.MarshalJson(fileModel))
	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.sig -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseSigModelByOpenssl(): len(results):", len(results))

	//get file hash
	sigModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	// get sig hex
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

	err = ParseSigModelByOpensslResults(results, sigModel)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl():ParseSigModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	sigModel.EContentType, err = ParseSigEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl():ParseSigEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	sigModel.SignerInfoModel, err = ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in mft
	eeCerFile, fileByte, eeCertStart, eeCertEnd, err := ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(eeCerFile)
	belogs.Debug("ParseSigModelByOpenssl():ParseByOpensslAns1ToX509:", eeCerFile, fileByte, eeCertStart, eeCertEnd)

	results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseSigModelByOpenssl(): len(results):", len(results))

	sigModel.Aki, sigModel.Ski, err = ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA, no SIA
	sigModel.AiaModel, _, err = ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	sigModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, eeCertStart, eeCertEnd)
	if err != nil {
		belogs.Error("ParseSigModelByOpenssl(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	belogs.Debug("ParseSigModelByOpenssl(): sigModel:", sigModel.String(), "  time(s):", time.Since(start))
	return nil
}
