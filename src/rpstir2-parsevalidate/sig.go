package parsevalidate

import (
	model "rpstir2-model"
	openssl "rpstir2-parsevalidate-openssl"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
)

//Try to store the error in statemode instead of returning err
func ParseValidateSig(certFile string) (sigModel model.SigModel, stateModel model.StateModel, err error) {

	stateModel = model.NewStateModel()
	err = parseSigModel(certFile, &sigModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateSig():parseSigModel err:", certFile, err)
		return sigModel, stateModel, nil
	}
	belogs.Debug("ParseValidateSig(): sigModel:", jsonutil.MarshalJson(sigModel))

	err = validateSigModel(&sigModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateSig():validateSigModel err:", certFile, err)
		return sigModel, stateModel, nil
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("ParseValidateSig():stateModel have errors or warnings", certFile, "     stateModel:", jsonutil.MarshalJson(stateModel))
	}

	belogs.Debug("ParseValidateSig(): sigModel.FilePath, sigModel.FileName, sigModel.Ski, sigModel.Aki:",
		sigModel.FilePath, sigModel.FileName, sigModel.Ski, sigModel.Aki)
	return sigModel, stateModel, nil
}

// some parse may return err, will stop
func parseSigModel(certFile string, sigModel *model.SigModel, stateModel *model.StateModel) error {
	belogs.Debug("parseSigModel(): certFile: ", certFile)
	sigModel.FilePath, sigModel.FileName = osutil.GetFilePathAndFileName(certFile)
	sigModel.Version = 0 //default
	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.sig -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(certFile)
	if err != nil {
		belogs.Error("parseSigModel(): GetResultsByOpensslAns1: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseSigModel(): len(results):", len(results))

	//get file hash
	sigModel.FileHash, err = hashutil.Sha256File(certFile)
	if err != nil {
		belogs.Error("parseSigModel(): Sha256File: err: ", err, ": "+certFile)
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

	err = openssl.ParseSigModelByOpensslResults(results, sigModel)
	if err != nil {
		belogs.Error("parseSigModel():ParseSigModelByOpensslResults  certFile:", certFile, "  err:", err, " will try parseMftModelByPacket")
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	sigModel.EContentType, err = openssl.ParseSigEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("parseSigModel():ParseSigEContentTypeByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	sigModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseSigModel():ParseSignerInfoModelByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in mft
	cerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(certFile, results)
	if err != nil {
		belogs.Error("parseSigModel():ParseByOpensslAns1ToX509  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(cerFile)
	belogs.Debug("parseSigModel():ParseByOpensslAns1ToX509:", cerFile, fileByte, start, end)

	results, err = opensslutil.GetResultsByOpensslX509(cerFile.Name())
	if err != nil {
		belogs.Error("parseSigModel(): GetResultsByOpensslX509: err: ", err, ": "+cerFile.Name())
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseSigModel(): len(results):", len(results))

	sigModel.Aki, sigModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA, no SIA
	sigModel.AiaModel, _, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseSigModel(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	sigModel.EeCertModel, err = ParseEeCertModel(cerFile.Name(), fileByte, start, end)
	if err != nil {
		belogs.Error("parseSigModel(): ParseEeCertModel: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	belogs.Debug("parseSigModel(): sigModel:", jsonutil.MarshalJson(sigModel))
	return nil
}

func validateSigModel(sigModel *model.SigModel, stateModel *model.StateModel) (err error) {
	return
}
