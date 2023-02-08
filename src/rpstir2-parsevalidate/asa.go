package parsevalidate

import (
	"errors"
	"strings"
	"time"

	model "rpstir2-model"
	openssl "rpstir2-parsevalidate-openssl"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
)

//Try to store the error in statemode instead of returning err
func ParseValidateAsa(certFile string) (asaModel model.AsaModel, stateModel model.StateModel, err error) {

	stateModel = model.NewStateModel()
	err = parseAsaModel(certFile, &asaModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateAsa():parseAsaModel err:", certFile, err)
		return asaModel, stateModel, nil
	}
	belogs.Debug("ParseValidateAsa(): asaModel:", jsonutil.MarshalJson(asaModel))

	err = validateAsaModel(&asaModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateAsa():validateAsaModel err:", certFile, err)
		return asaModel, stateModel, nil
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("ParseValidateAsa():stateModel have errors or warnings", certFile, "     stateModel:", jsonutil.MarshalJson(stateModel))
	}

	belogs.Debug("ParseValidateAsa(): asaModel.FilePath, asaModel.FileName, asaModel.Ski, asaModel.Aki:",
		asaModel.FilePath, asaModel.FileName, asaModel.Ski, asaModel.Aki)
	return asaModel, stateModel, nil
}

// some parse may return err, will stop
func parseAsaModel(certFile string, asaModel *model.AsaModel, stateModel *model.StateModel) error {
	belogs.Debug("parseAsaModel(): certFile: ", certFile)
	fileLength, err := fileutil.GetFileLength(certFile)
	if err != nil {
		belogs.Error("parseAsaModel(): GetFileLength: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	} else if fileLength == 0 {
		belogs.Error("parseAsaModel(): GetFileLength, fileLenght is emtpy: " + certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + certFile + " is empty")
	}

	asaModel.FilePath, asaModel.FileName = osutil.GetFilePathAndFileName(certFile)
	asaModel.Version = 0 //default
	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.sig -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(certFile)
	if err != nil {
		belogs.Error("parseAsaModel(): GetResultsByOpensslAns1: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseAsaModel(): len(results):", len(results))

	//get file hash
	asaModel.FileHash, err = hashutil.Sha256File(certFile)
	if err != nil {
		belogs.Error("parseAsaModel(): Sha256File: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

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

	err = openssl.ParseAsaModelByOpensslResults(results, asaModel)
	if err != nil {
		belogs.Error("parseAsaModel():ParseSigModelByOpensslResults  certFile:", certFile, "  err:", err, " will try parseMftModelByPacket")
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	asaModel.EContentType, err = openssl.ParseAsaEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("parseAsaModel():ParseAsaEContentTypeByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	asaModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseAsaModel():ParseSignerInfoModelByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in mft
	cerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(certFile, results)
	if err != nil {
		belogs.Error("parseAsaModel():ParseByOpensslAns1ToX509  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(cerFile)
	belogs.Debug("parseAsaModel():ParseByOpensslAns1ToX509:", cerFile, fileByte, start, end)

	results, err = opensslutil.GetResultsByOpensslX509(cerFile.Name())
	if err != nil {
		belogs.Error("parseAsaModel(): GetResultsByOpensslX509: err: ", err, ": "+cerFile.Name())
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseAsaModel(): len(results):", len(results))

	asaModel.Aki, asaModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA,  SIA
	asaModel.AiaModel, asaModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseAsaModel(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	asaModel.EeCertModel, err = ParseEeCertModel(cerFile.Name(), fileByte, start, end)
	if err != nil {
		belogs.Error("parseAsaModel(): ParseEeCertModel: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	belogs.Debug("parseAsaModel(): asaModel:", jsonutil.MarshalJson(asaModel))
	return nil
}

func validateAsaModel(asaModel *model.AsaModel, stateModel *model.StateModel) (err error) {
	return
}

func updateAsaByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireAsaDb(now)
	if err != nil {
		belogs.Error("updateAsaByCheckAll(): getExpireAsaDb:  err: ", err)
		return err
	}
	belogs.Info("updateAsaByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("updateAsaByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NotAfter of EE is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter of EE is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleEe") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("updateAsaByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateAsaStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("updateAsaByCheckAll(): updateAsaStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("updateAsaByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
