package parsevalidate

import (
	"errors"
	"strconv"
	"strings"
	"time"

	model "rpstir2-model"
	openssl "rpstir2-parsevalidate-openssl"

	"github.com/cpusoft/goutil/asn1util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/regexputil"
)

//Try to store the error in statemode instead of returning err
func ParseValidateCrl(certFile string) (crlModel model.CrlModel, stateModel model.StateModel, err error) {
	stateModel = model.NewStateModel()
	err = parseCrlModel(certFile, &crlModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateCrl():parseCrlByOpenssl err:", err)
		return crlModel, stateModel, nil
	}

	err = validateCrlModel(&crlModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateCrl():validateCrlModel err:", certFile, err)
		return crlModel, stateModel, nil
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("ParseValidateCrl():stateModel have errors or warnings", certFile, "     stateModel:", jsonutil.MarshalJson(stateModel))
	}

	belogs.Debug("ParseValidateCrl():  crlModel.FilePath, crlModel.FileName, crlModel.Aki:",
		crlModel.FilePath, crlModel.FileName, crlModel.Aki)
	return crlModel, stateModel, nil
}

// some parse may return err, will stop
func parseCrlModel(certFile string, crlModel *model.CrlModel, stateModel *model.StateModel) error {
	belogs.Debug("parseCrlModel():certFile: ", certFile)
	fileLength, err := fileutil.GetFileLength(certFile)
	if err != nil {
		belogs.Error("parseCrlModel(): GetFileLength: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	} else if fileLength == 0 {
		belogs.Error("parseCrlModel(): GetFileLength, fileLenght is emtpy: " + certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + certFile + " is empty")
	}

	crlModel.FilePath, crlModel.FileName = osutil.GetFilePathAndFileName(certFile)

	fileByte, fileDecodeBase64Byte, err := asn1util.ReadFileAndDecodeBase64(certFile)
	if err != nil {
		belogs.Error("parseCrlModel():ReadFile return err: ", certFile, err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	//get file hash
	crlModel.FileHash = hashutil.Sha256(fileByte)

	err = openssl.ParseCrlModelByX509(fileDecodeBase64Byte, crlModel)
	if err != nil {
		belogs.Error("parseCrlModel():ParseCrlModelByX509 err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	results, err := opensslutil.GetResultsByOpensslAns1(certFile)
	if err != nil {
		belogs.Error("parseCrlModel(): GetResultsByOpensslAns1: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseCrlModel(): len(results):", len(results))

	err = openssl.ParseCrlModelByOpensslResults(results, crlModel)
	if err != nil {
		belogs.Error("parseCrlModel(): ParseCrlModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

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

func updateCrlByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireCrlDb(now)
	if err != nil {
		belogs.Error("updateCrlByCheckAll(): getExpireCrlDb:  err: ", err)
		return err
	}
	belogs.Info("updateCrlByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("updateCrlByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NextUpdate is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleCrl") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("updateCrlByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateCrlStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("updateCrlByCheckAll(): updateCrlStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("updateCrlByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
