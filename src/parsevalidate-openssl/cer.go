package openssl

import (
	"crypto/x509"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
)

func ParseCerModelByOpenssl(fileModel *model.FileModel, cerModel *model.CerModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("ParseCerModelByOpenssl():fileModel ", fileModel)
	//get file byte
	_, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	err = ParseCerModelByX509(fileDecodeBase64Byte, cerModel)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl():ParseCerModelByX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	// results will be used
	results, err := opensslutil.GetResultsByOpensslX509(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): GetResultsByOpensslX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseCerModelByOpenssl(): GetResultsByOpensslX509 len(results):", len(results))

	// IP
	var noCerIpAddress bool
	cerModel.CerIpAddressModel, noCerIpAddress, err = ParseCerIpAddressModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseCerIpAddressModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain IP address",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		//no return
		//return
	}

	// AS
	var noAsn bool
	cerModel.AsnModel, noAsn, err = ParseAsnModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseAsnModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain ASN",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	//P3277 rescert_ip_asnum_chk
	//P3086 rescert_ip_resources_chk   RFC6487 4.8.10.  IP Resources
	if noCerIpAddress && noAsn {
		belogs.Error("ParseCerModelByOpenssl(): noCerIpAddress && noAsn, fileModel:", jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to find INR extension",
			Detail: "There is neither address nor ASN number"}
		stateModel.AddError(&stateMsg)
	}

	// AIA SIA
	cerModel.AiaModel, cerModel.SiaModel, err = ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain AIA or SIA",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	//  SignatureInnerAlgorithm  SignatureOuterAlgorithm  PublicKeyAlgorithm
	cerModel.SignatureInnerAlgorithm, cerModel.SignatureOuterAlgorithm, cerModel.PublicKeyAlgorithm, err =
		ParseSignatureAndPublicKeyByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseSignatureAndPublicKeyByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain Signature Algorithm",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	//keyusage ,critical
	cerModel.KeyUsageModel.Critical, cerModel.KeyUsageModel.KeyUsageValue, err = ParseKeyUsageModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseKeyUsageModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain Key Usage",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// crldp critical
	cerModel.CrldpModel.Critical, err = ParseCrldpModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseCrldpModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain CRL Distribution Points",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// cert policy:CP
	cerModel.CertPolicyModel, err = ParseCertPolicyModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseCertPolicyModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain Certificate Policies",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// basic contraints : critical
	cerModel.BasicConstraintsModel.Critical, err = ParseBasicConstraintsModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerModelByOpenssl(): ParseBasicConstraintsModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to obtain Basic Constraints",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	belogs.Debug("ParseCerModelByOpenssl(): cerModel:", cerModel.String(), "  time(s):", time.Since(start))
	return nil
}

func ParseCerSimpleModelByOpenssl(fileModel *model.FileModel) (parseCerSimple model.ParseCerSimple, err error) {
	// results will be used
	belogs.Debug("ParseCerSimpleModelByOpenssl(): fileModel:", jsonutil.MarshalJson(fileModel))
	results, err := opensslutil.GetResultsByOpensslX509(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCerSimpleModelByOpenssl(): GetResultsByOpensslX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return parseCerSimple, err
	}
	belogs.Debug("ParseCerSimpleModelByOpenssl(): GetResultsByOpensslX509, fileModel:", jsonutil.MarshalJson(fileModel), "  len(results):", len(results))

	//  SIA
	_, siaModel, err := ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseCerSimpleModelByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return parseCerSimple, err
	}
	belogs.Debug("ParseCerSimpleModelByOpenssl():ParseAiaModelSiaModelByOpensslResults, fileModel:", jsonutil.MarshalJson(fileModel), " siaModel:", siaModel)
	parseCerSimple.RpkiNotify = siaModel.RpkiNotify
	parseCerSimple.CaRepository = siaModel.CaRepository

	// get SubjectPublicKeyInfo
	_, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCerSimpleModelByOpenssl(): ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return parseCerSimple, err
	}
	cer, err := x509.ParseCertificate(fileDecodeBase64Byte)
	if err != nil {
		belogs.Error("ParseCerSimpleModelByOpenssl(): ParseCertificate fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return parseCerSimple, err
	}
	parseCerSimple.SubjectPublicKeyInfo = cer.RawSubjectPublicKeyInfo
	return parseCerSimple, nil
}
