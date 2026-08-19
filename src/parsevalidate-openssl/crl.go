package openssl

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
)

func ParseCrlModelByOpenssl(fileModel *model.FileModel, crlModel *model.CrlModel, stateModel *model.StateModel) error {
	start := time.Now()
	belogs.Debug("ParseCrlModelByOpenssl():fileModel:", jsonutil.MarshalJson(fileModel))

	_, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCrlModelByOpenssl():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	err = ParseCrlModelByX509(fileDecodeBase64Byte, crlModel)
	if err != nil {
		belogs.Error("ParseCrlModelByOpenssl():ParseCrlModelByX509 err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCrlModelByOpenssl(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseCrlModelByOpenssl(): GetResultsByOpensslAns1 fileModel:", jsonutil.MarshalJson(fileModel), "  len(results):", len(results))

	err = ParseCrlModelByOpensslResults(results, crlModel)
	if err != nil {
		belogs.Error("ParseCrlModelByOpenssl(): ParseCrlModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseCrlModelByOpenssl(): crlModel:", crlModel.String(), "  time(s):", time.Since(start))
	return nil
}
