package parsevalidatecore

import (
	"errors"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
)

/*
MFT: Manifests for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6486/?include_text=1

ROA: A Profile for Route Origin Authorizations (ROAs)
https://datatracker.ietf.org/doc/rfc6482/?include_text=1

CRL: Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
https://datatracker.ietf.org/doc/rfc5280/?include_text=1

EE: Signed Object Template for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6488/?include_text=1

CER: IP/AS:  X.509 Extensions for IP Addresses and AS Identifiers
https://datatracker.ietf.org/doc/rfc3779/?include_text=1

CER: A Profile for X.509 PKIX Resource Certificates
https://datatracker.ietf.org/doc/rfc6487/?include_text=1



A Profile for X.509 PKIX Resource Certificates
https://datatracker.ietf.org/doc/rfc6487/?include_text=1


A Profile for Route Origin Authorizations (ROAs)
https://datatracker.ietf.org/doc/rfc6482/?include_text=1

Signed Object Template for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6488/?include_text=1

X.509 Extensions for IP Addresses and AS Identifiers
https://datatracker.ietf.org/doc/rfc3779/?include_text=1


Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
https://datatracker.ietf.org/doc/rfc5280/?include_text=1
*/
// upload file to parse

func ParseValidateFile(certFile string) (certType string, certModel interface{},
	stateModel model.StateModel, originModel model.OriginModel, fileHash string, err error) {
	fileModel := &model.FileModel{
		FilePathName:     certFile,
		TempFilePathName: certFile,
	}
	return ParseValidateFileModel(fileModel)
}

func ParseValidateFileModel(fileModel *model.FileModel) (certType string, certModel interface{},
	stateModel model.StateModel, originModel model.OriginModel, fileHash string, err error) {
	start := time.Now()
	belogs.Debug("ParseValidateFileModel(): parsevalidate start, fileModel:", jsonutil.MarshalJson(fileModel))
	fileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseValidateFileModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return "", nil, stateModel, originModel, "", err
	}
	belogs.Debug("ParseValidateFileModel(): Sha256File, fileModel:", jsonutil.MarshalJson(fileModel), " len(fileHash):", len(fileHash))

	originModel = *model.JudgeOriginByFilePath(fileModel.FilePathName)
	belogs.Debug("ParseValidateFileModel(): JudgeOriginByFilePath, fileModel:", jsonutil.MarshalJson(fileModel), " originModel:", jsonutil.MarshalJson(originModel))

	if strings.HasSuffix(fileModel.FilePathName, ".cer") {
		cerModel, stateModel, err := parseValidateCer(fileModel)
		belogs.Debug("ParseValidateFileModel(): after parseValidateCer():fileModel:", jsonutil.MarshalJson(fileModel), "  cerModel:", cerModel.String(),
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err, "  time(s):", time.Since(start))
		return "cer", cerModel, stateModel, originModel, fileHash, err
	} else if strings.HasSuffix(fileModel.FilePathName, ".crl") {
		crlModel, stateModel, err := parseValidateCrl(fileModel)
		belogs.Debug("ParseValidateFileModel(): after parseValidateCrl(): fileModel:", jsonutil.MarshalJson(fileModel), "  crlModel:", crlModel.String(),
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err, "  time(s):", time.Since(start))
		return "crl", crlModel, stateModel, originModel, fileHash, err
	} else if strings.HasSuffix(fileModel.FilePathName, ".mft") {
		mftModel, stateModel, err := parseValidateMft(fileModel)
		belogs.Debug("ParseValidateFileModel(): after parseValidateMft():fileModel:", jsonutil.MarshalJson(fileModel), "  mftModel:", mftModel.String(),
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err, "  time(s):", time.Since(start))
		return "mft", mftModel, stateModel, originModel, fileHash, err
	} else if strings.HasSuffix(fileModel.FilePathName, ".roa") {
		roaModel, stateModel, err := parseValidateRoa(fileModel)
		belogs.Debug("ParseValidateFileModel(): after parseValidateRoa(): fileModel:", jsonutil.MarshalJson(fileModel), "  roaModel:", roaModel.String(),
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err, "  time(s):", time.Since(start))
		return "roa", roaModel, stateModel, originModel, fileHash, err
	} else if strings.HasSuffix(fileModel.FilePathName, ".asa") {
		asaModel, stateModel, err := parseValidateAsa(fileModel)
		belogs.Debug("ParseValidateFileModel(): after parseValidateAsa(): fileModel:", jsonutil.MarshalJson(fileModel), "  asaModel:", asaModel.String(),
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err, "  time(s):", time.Since(start))
		return "asa", asaModel, stateModel, originModel, fileHash, err
	} else if strings.HasSuffix(fileModel.FilePathName, ".moa") {
		moaModel, stateModel, err := parseValidateMoa(fileModel)
		belogs.Debug("ParseValidateFileModel(): after parseValidateMoa(): fileModel:", jsonutil.MarshalJson(fileModel), "  moaModel:", moaModel.String(),
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err, "  time(s):", time.Since(start))
		return "moa", moaModel, stateModel, originModel, fileHash, err
	} else if strings.HasSuffix(fileModel.FilePathName, ".gbr") {
		belogs.Debug("ParseValidateFileModel(): not support .gbr file: fileModel:", jsonutil.MarshalJson(fileModel))
		return "", nil, stateModel, originModel, "", errors.New("gbr file type is not supported")
	} else {
		return "", nil, stateModel, originModel, "", errors.New("unknown file type")
	}
}

// only parse cer to get ca repository/rpkiNotify, raw subjct public key info
func ParseFileSimple(certFile string) (parseCerSimple model.ParseCerSimple, err error) {
	if strings.HasSuffix(certFile, ".cer") {
		fileModel := &model.FileModel{
			FilePathName:     certFile,
			TempFilePathName: certFile,
		}
		return parseCerSimpleModel(fileModel)
	}
	return parseCerSimple, errors.New("unknown file type")
}
