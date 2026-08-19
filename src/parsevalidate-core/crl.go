package parsevalidatecore

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	asn1 "github.com/bgpsecurity/rpstir2/parsevalidate-asn1"
	openssl "github.com/bgpsecurity/rpstir2/parsevalidate-openssl"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/regexputil"
)

func parseValidateCrl(fileModel *model.FileModel) (crlModel model.CrlModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateCrl(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseCrlModel(fileModel, &crlModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateCrl():parseCrlModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return crlModel, stateModel, err
	}
	belogs.Debug("parseValidateCrl(): crlModel:", crlModel.String(), "  time(s):", time.Since(start))

	err = validateCrlModel(&crlModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateCrl():validateCrlModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return crlModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateCrl():stateModel have errors or warnings, fileModel:", jsonutil.MarshalJson(fileModel), "   stateModel:", jsonutil.MarshalJson(stateModel))
	}
	stateModel.JudgeState()
	belogs.Debug("parseValidateCrl():  crlModel.FilePath, crlModel.FileName, crlModel.Aki:",
		crlModel.FilePath, crlModel.FileName, crlModel.Aki, "  time(s):", time.Since(start))
	return crlModel, stateModel, nil
}

// some parse may return err, will stop
func parseCrlModel(fileModel *model.FileModel, crlModel *model.CrlModel, stateModel *model.StateModel) error {
	belogs.Debug("parseCrlModel():fileModel:", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseCrlModel(): GetFileLength fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseCrlModel(): fileLength:", fileLength, " tempFile:", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseCrlModel(): GetFileLength fail, fileLenght is empty, fileModel:",
			jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	crlModel.FilePath, crlModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	//get file hash
	crlModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseCrlModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	err = asn1.ParseCrlModelByAsn1(fileModel, crlModel, stateModel)
	if err != nil {
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("parseCrlModel(): ParseCrlModelByOpenssl fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("parseCrlModel(): ParseCrlModelByOpenssl fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		// have set stateMsg in ParseCrlModelByAsn1 .
		err = openssl.ParseCrlModelByOpenssl(fileModel, crlModel, stateModel)
		if err != nil {
			belogs.Error("parseCrlModel(): ParseCrlModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			// have set stateMsg in ParseCrlModelByOpenssl .
			return err
		}
	}

	/*
		fileByte, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseCrlModel():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to read file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

		err = openssl.ParseCrlModelByX509(fileDecodeBase64Byte, crlModel)
		if err != nil {
			belogs.Error("parseCrlModel():ParseCrlModelByX509 err:", err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

		results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseCrlModel(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseCrlModel(): GetResultsByOpensslAns1 fileModel:", jsonutil.MarshalJson(fileModel), "  len(results):", len(results))

		err = openssl.ParseCrlModelByOpensslResults(results, crlModel)
		if err != nil {
			belogs.Error("parseCrlModel(): ParseCrlModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
	*/
	belogs.Debug("parseCrlModel(): crlModel:", crlModel.String())
	return nil
}

// https://datatracker.ietf.org/doc/rfc5280/?include_text=1  Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
// https://datatracker.ietf.org/doc/rfc6487/?include_text=1   5.Resource Certificate Revocation Lists
// rpstir:sqlh.c P3098 add_crl() ;  P4556 crl_profile_chk();
// TODO P1727 verify_crl(), need use x508 to check crl;
// TODO P4349 revoke_cert_by_serial() actually to revoke cer file
func validateCrlModel(crlModel *model.CrlModel, stateModel *model.StateModel) (err error) {

	if crlModel.Version != 1 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Wrong Version number",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if crlModel.TbsAlgorithm != "sha256WithRSAEncryption" {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Algorithm is not sha256WithRSAEncryption",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if crlModel.CertAlgorithm != "sha256WithRSAEncryption" {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Algorithm is not sha256WithRSAEncryption",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(crlModel.IssuerAll) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Issuer is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	//check time
	now := time.Now()
	if crlModel.ThisUpdate.IsZero() {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if crlModel.NextUpdate.IsZero() {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	//thisUpdate precedes nextUpdate.
	if crlModel.ThisUpdate.After(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate is later than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", thisUpdate is " + convert.Time2StringZone(crlModel.ThisUpdate)}
		if conf.Bool("policy::allowNotYetCrl") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}
	}
	if crlModel.NextUpdate.Before(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(crlModel.NextUpdate)}
		if conf.Bool("policy::allowStaleCrl") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}
	}
	if crlModel.ThisUpdate.After(crlModel.NextUpdate) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than ThisUpdate",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// crl number , max length is 20
	if crlModel.CrlNumber == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "CRL Number is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	} else if len(strconv.FormatInt(int64(crlModel.CrlNumber), 10)) > 20 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "CRL Number is too long",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	if len(crlModel.Aki) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(crlModel.Aki) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	for _, one := range crlModel.RevokedCertModels {
		if one.RevocationTime.IsZero() {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "One revocation times in the revocation list is empty",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if len(one.Sn) == 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "One SN in the revocation list is empty",
				Detail: one.Sn}
			stateModel.AddError(&stateMsg)
		} else {
			if len(one.Sn) > 20*2 {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "One SN length in the revocation list is wrong",
					Detail: one.Sn}
				stateModel.AddError(&stateMsg)
			}
			isHex, err := regexputil.IsHex(one.Sn)
			if !isHex || err != nil {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "One SN in the revocation list is not a hexadecimal number",
					Detail: one.Sn}
				stateModel.AddError(&stateMsg)
			}
		}
		if len(one.Extensions) > 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The Extensions is not empty",
				Detail: jsonutil.MarshalJson(one.Extensions)}
			stateModel.AddError(&stateMsg)
		}

	}
	belogs.Debug("validateCrlModel():filePath, fileName,stateModel:",
		crlModel.FilePath, crlModel.FileName, jsonutil.MarshalJson(stateModel))
	return nil
}
