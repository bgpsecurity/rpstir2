package openssl

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
)

// some parse may return err, will stop
func ParseAsaModelByOpenssl(fileModel *model.FileModel, asaModel *model.AsaModel, stateModel *model.StateModel) error {
	start := time.Now()
	belogs.Debug("ParseAsaModelByOpenssl(): fileModel:", jsonutil.MarshalJson(fileModel))

	asaModel.Version = 0 //default
	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.sig -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseAsaModelByOpenssl(): len(results):", len(results))

	// get asa hex
	// first HEX DUMP
	/*
	   23:d=3  hl=2 l=   1 prim: INTEGER           :03
	   26:d=3  hl=2 l=  13 cons: SET
	   28:d=4  hl=2 l=  11 cons: SEQUENCE
	   30:d=5  hl=2 l=   9 prim: OBJECT            :sha256
	   41:d=3  hl=2 l=  55 cons: SEQUENCE
	   43:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.49
	   56:d=4  hl=2 l=  40 cons: cont [ 0 ]
	   58:d=5  hl=2 l=  38 prim: OCTET STRING      [HEX DUMP]:30240203033979301D3005020300FDE83009020300FDE9040200013009020300FDEA04020002
	*/

	err = ParseAsaModelByOpensslResults(results, asaModel)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl():ParseAsaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	asaModel.EContentType, err = ParseAsaEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl():ParseAsaEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	asaModel.SignerInfoModel, err = ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in mft
	eeCerFile, fileByte, eeCertStart, eeCertEnd, err := ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(eeCerFile)
	belogs.Debug("ParseAsaModelByOpenssl():ParseByOpensslAns1ToX509, fileModel:", jsonutil.MarshalJson(fileModel), "  len(fileByte):", fileByte,
		"  eeCertStart:", eeCertStart, "   eeCertEnd:", eeCertEnd)

	results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseAsaModelByOpenssl(): len(results):", len(results))

	asaModel.Aki, asaModel.Ski, err = ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA,  SIA
	asaModel.AiaModel, asaModel.SiaModel, err = ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	asaModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, eeCertStart, eeCertEnd)
	if err != nil {
		belogs.Error("ParseAsaModelByOpenssl(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	belogs.Debug("ParseAsaModelByOpenssl(): asaModel:", asaModel.String(), "  time(s):", time.Since(start))
	return nil
}
