package parsevalidatecore

import (
	"errors"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	asn1 "github.com/bgpsecurity/rpstir2/parsevalidate-asn1"
	openssl "github.com/bgpsecurity/rpstir2/parsevalidate-openssl"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func parseValidateAsa(fileModel *model.FileModel) (asaModel model.AsaModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateAsa(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseAsaModel(fileModel, &asaModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateAsa():parseAsaModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return asaModel, stateModel, err
	}
	belogs.Debug("parseValidateAsa(): asaModel:", asaModel.String(), "  time(s):", time.Since(start))

	err = validateAsaModel(&asaModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateAsa():validateAsaModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return asaModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateAsa():stateModel have errors or warnings, fileModel:", jsonutil.MarshalJson(fileModel), "     stateModel:", jsonutil.MarshalJson(stateModel))
	}

	stateModel.JudgeState()
	belogs.Debug("parseValidateAsa(): asaModel.FilePath, asaModel.FileName, asaModel.Ski, asaModel.Aki:",
		asaModel.FilePath, asaModel.FileName, asaModel.Ski, asaModel.Aki, "  time(s):", time.Since(start))
	return asaModel, stateModel, nil
}

// some parse may return err, will stop
func parseAsaModel(fileModel *model.FileModel, asaModel *model.AsaModel, stateModel *model.StateModel) error {
	belogs.Debug("parseAsaModel(): fileModel:", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseAsaModel(): GetFileLength fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseAsaModel(): fileLength:", fileLength, " tempFile:", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseAsaModel(): GetFileLength fail, fileLenght is empty, fileModel:",
			jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	asaModel.FilePath, asaModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	//get file hash
	asaModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseAsaModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	asaModel.Version = 0 //default
	err = asn1.ParseAsaModelByAsn1(fileModel, asaModel, stateModel)
	if err != nil {
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("parseAsaModel(): ParseAsaModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("parseAsaModel(): ParseAsaModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		// have set stateMsg in ParseAsaModelByAsn1 here.
		// again by openssl
		err = openssl.ParseAsaModelByOpenssl(fileModel, asaModel, stateModel)
		if err != nil {
			belogs.Error("parseAsaModel(): ParseAsaModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			// have set stateMsg in ParseAsaModelByOpenssl here.
			return err
		}
	}

	/*
		//https://blog.csdn.net/Zhymax/article/details/7683925
		//openssl asn1parse -in -ard.sig -inform DER
		results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseAsaModel(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseAsaModel(): len(results):", len(results))

		// get asa hex
		// first HEX DUMP

		   23:d=3  hl=2 l=   1 prim: INTEGER           :03
		   26:d=3  hl=2 l=  13 cons: SET
		   28:d=4  hl=2 l=  11 cons: SEQUENCE
		   30:d=5  hl=2 l=   9 prim: OBJECT            :sha256
		   41:d=3  hl=2 l=  55 cons: SEQUENCE
		   43:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.49
		   56:d=4  hl=2 l=  40 cons: cont [ 0 ]
		   58:d=5  hl=2 l=  38 prim: OCTET STRING      [HEX DUMP]:30240203033979301D3005020300FDE83009020300FDE9040200013009020300FDEA04020002


		err = openssl.ParseAsaModelByOpensslResults(results, asaModel)
		if err != nil {
			belogs.Error("parseAsaModel():ParseAsaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		asaModel.EContentType, err = openssl.ParseAsaEContentTypeByOpensslResults(results)
		if err != nil {
			belogs.Error("parseAsaModel():ParseAsaEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		asaModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseAsaModel():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// get cer info in mft
		eeCerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
		if err != nil {
			belogs.Error("parseAsaModel():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		defer osutil.CloseAndRemoveFile(eeCerFile)
		belogs.Debug("parseAsaModel():ParseByOpensslAns1ToX509, fileModel:", jsonutil.MarshalJson(fileModel), "  len(fileByte):", fileByte,
			"  start:", start, "   end:", end)

		results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
		if err != nil {
			belogs.Error("parseAsaModel(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseAsaModel(): len(results):", len(results))

		asaModel.Aki, asaModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// AIA,  SIA
		asaModel.AiaModel, asaModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseAsaModel(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		asaModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, start, end)
		if err != nil {
			belogs.Error("parseAsaModel(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
	*/
	belogs.Debug("parseAsaModel(): asaModel:", asaModel.String())
	return nil
}

func validateAsaModel(asaModel *model.AsaModel, stateModel *model.StateModel) (err error) {
	if asaModel.Version != 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Wrong Version number",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	if len(asaModel.CustomerAsns) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "CustomerAsns is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// ski aki
	if len(asaModel.Ski) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(asaModel.Ski) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(asaModel.Aki) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(asaModel.Aki) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	return
}
