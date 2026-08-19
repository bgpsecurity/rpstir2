package parsevalidatecore

import (
	"errors"
	"strings"
	"time"

	model "github.com/bgpsecurity/rpstir2/model"
	asn1 "github.com/bgpsecurity/rpstir2/parsevalidate-asn1"
	openssl "github.com/bgpsecurity/rpstir2/parsevalidate-openssl"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func parseValidateMoa(fileModel *model.FileModel) (moaModel model.MoaModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateMoa(): will parse", "fileModel", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseMoaModel(fileModel, &moaModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateMoa(): parseMoaModel fail",
			"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		return moaModel, stateModel, err
	}
	belogs.Debug("parseValidateMoa(): get moaModel",
		"moaModel", moaModel.String(), "time(s)", time.Since(start))

	err = validateMoaModel(&moaModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateMoa():validateMoaModel fail",
			"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		return moaModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateMoa():stateModel have errors or warnings",
			"fileModel", jsonutil.MarshalJson(fileModel), "stateModel", jsonutil.MarshalJson(stateModel))
	}
	stateModel.JudgeState()
	belogs.Info("parseValidateMoa(): get moaModel ok",
		"moaModel.FilePath", moaModel.FilePath,
		"moaModel.FileName", moaModel.FileName, "moaModel.Ski", moaModel.Ski,
		"moaModel.Aki", moaModel.Aki, "time(s)", time.Since(start))
	return moaModel, stateModel, nil
}

func parseMoaModel(fileModel *model.FileModel, moaModel *model.MoaModel, stateModel *model.StateModel) error {
	belogs.Debug("parseMoaModel(): input param", "fileModel", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseMoaModel(): GetFileLength fail",
			"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseMoaModel(): GetFileLength ok",
		"fileLength", fileLength, "tempFile", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseMoaModel(): GetFileLength fail, fileLength is empty",
			"fileModel", jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	moaModel.FilePath, moaModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	//get file hash
	moaModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseMoaModel(): Sha256File fail",
			"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	err = asn1.ParseMoaModelByAsn1(fileModel, moaModel, stateModel)
	if err != nil {
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Error("parseMoaModel(): ParseMoaModelByAsn1 fail, indefinite length found, will parse again by openssl",
				"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		} else {
			belogs.Error("parseMoaModel(): ParseMoaModelByAsn1 fail, will parse again by openssl",
				"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		} // have set stateMsg in ParseMoaModelByAsn1 .
		err = openssl.ParseMoaModelByOpenssl(fileModel, moaModel, stateModel)
		if err != nil {
			belogs.Error("parseMoaModel(): ParseMoaModelByOpenssl fail",
				"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
			// have set stateMsg in ParseMoaModelByOpenssl .
			return err
		}
	}

	belogs.Debug("parseMoaModel(): ok", "moaModel", moaModel.String())
	return nil
}

func validateMoaModel(moaModel *model.MoaModel, stateModel *model.StateModel) (err error) {

	if moaModel.Version != 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Wrong Version number",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	if moaModel.Ipv6MappingPrefix == "" {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Ipv6MappingPrefix is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(moaModel.Ipv4Prefixes) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Ipv4Prefixes is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// ski aki
	if len(moaModel.Ski) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(moaModel.Ski) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(moaModel.Aki) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(moaModel.Aki) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	/*
		// check time
		 ValidateEeCertModel(stateModel, &moaModel.EeModel)
		 ValidateSignerInfoModel(stateModel, &moaModel.SignerInfoModel)

	*/
	belogs.Info("validateMoaModel(): ok stateModel",
		"moaModel.FilePath", moaModel.FilePath, "moaModel.FileName", moaModel.FileName,
		"stateModel", jsonutil.MarshalJson(stateModel))

	return nil
}
