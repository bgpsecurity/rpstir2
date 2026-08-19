package parsevalidatecore

import (
	"errors"
	"fmt"
	"net/url"
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
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/regexputil"
)

func parseValidateCer(fileModel *model.FileModel) (cerModel model.CerModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateCer(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseCerModel(fileModel, &cerModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateCer():parseCerModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return cerModel, stateModel, err
	}
	belogs.Debug("parseValidateCer(): cerModel:", cerModel.String(), "  time(s):", time.Since(start))

	err = validateCerModel(&cerModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateCer():validateCerModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return cerModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateCer():stateModel have errors or warnings, fileModel:", jsonutil.MarshalJson(fileModel), "  stateModel:", jsonutil.MarshalJson(stateModel))
	}
	stateModel.JudgeState()
	belogs.Debug("parseValidateCer():cerModel.FilePath, cerModel.FileName, cerModel.Ski, cerModel.Aki:",
		cerModel.FilePath, cerModel.FileName, cerModel.Ski, cerModel.Aki, "  time(s):", time.Since(start))
	return cerModel, stateModel, nil
}

// some parse may return err, will stop
func parseCerModel(fileModel *model.FileModel, cerModel *model.CerModel, stateModel *model.StateModel) (err error) {

	belogs.Debug("parseCerModel(): fileModel:", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseCerModel(): GetFileLength fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseCerModel(): fileLength:", fileLength, " tempFile:", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseCerModel(): GetFileLength, fileLenght is empty, fail, fileModel:",
			jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	cerModel.FilePath, cerModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	cerModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseCerModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	err = asn1.ParseCerModelByAsn1(fileModel, false, cerModel, stateModel)
	if err != nil {
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("parseCerModel(): ParseCerModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("parseCerModel(): ParseCerModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		// have set stateMsg in ParseCerModelByAsn1 .
		err = openssl.ParseCerModelByOpenssl(fileModel, cerModel, stateModel)
		if err != nil {
			belogs.Error("parseCerModel(): ParseCerModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			// have set stateMsg in ParseCerModelByOpenssl .
			return err
		}
	}

	/*
		//get file byte
		fileByte, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseCerModel():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to read file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

		err = openssl.ParseCerModelByX509(fileDecodeBase64Byte, cerModel)
		if err != nil {
			belogs.Error("parseCerModel():ParseCerModelByX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

		// results will be used
		results, err := opensslutil.GetResultsByOpensslX509(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseCerModel(): GetResultsByOpensslX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseCerModel(): GetResultsByOpensslX509 len(results):", len(results))

		// IP
		var noCerIpAddress bool
		cerModel.CerIpAddressModel, noCerIpAddress, err = openssl.ParseCerIpAddressModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseCerIpAddressModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain IP address",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			//no return
			//return
		}

		// AS
		var noAsn bool
		cerModel.AsnModel, noAsn, err = openssl.ParseAsnModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseAsnModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain ASN",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		//P3277 rescert_ip_asnum_chk
		//P3086 rescert_ip_resources_chk   RFC6487 4.8.10.  IP Resources
		if noCerIpAddress && noAsn {
			belogs.Error("parseCerModel(): noCerIpAddress && noAsn, fileModel:", jsonutil.MarshalJson(fileModel))
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to find INR extension",
				Detail: "There is neither address nor ASN number"}
			stateModel.AddError(&stateMsg)
		}

		// AIA SIA
		cerModel.AiaModel, cerModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain AIA or SIA",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		//  SignatureInnerAlgorithm  SignatureOuterAlgorithm  PublicKeyAlgorithm
		cerModel.SignatureInnerAlgorithm, cerModel.SignatureOuterAlgorithm, cerModel.PublicKeyAlgorithm, err =
			openssl.ParseSignatureAndPublicKeyByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseSignatureAndPublicKeyByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain Signature Algorithm",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		//keyusage ,critical
		cerModel.KeyUsageModel.Critical, cerModel.KeyUsageModel.KeyUsageValue, err = openssl.ParseKeyUsageModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseKeyUsageModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain Key Usage",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// crldp critical
		cerModel.CrldpModel.Critical, err = openssl.ParseCrldpModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseCrldpModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain CRL Distribution Points",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// cert policy:CP
		cerModel.CertPolicyModel, err = openssl.ParseCertPolicyModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseCertPolicyModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain Certificate Policies",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// basic contraints : critical
		cerModel.BasicConstraintsModel.Critical, err = openssl.ParseBasicConstraintsModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseCerModel(): ParseBasicConstraintsModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to obtain Basic Constraints",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
	*/
	belogs.Debug("parseCerModel(): cerModel:", cerModel.String())
	return nil
}

// https://datatracker.ietf.org/doc/rfc6487/?include_text=1
// sqlh.c P3066 add_cert(): --> add_cert_2()
// myssl.c P3968 rescert_profile_chk
func validateCerModel(cerModel *model.CerModel, stateModel *model.StateModel) (err error) {

	// myssl.c  P1892 rescert_flags_chk TODO

	// version : RFC6487 4.1.  Version
	// myssl.c P1952 rescert_version_chk
	if cerModel.Version != 3 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Wrong Version number",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// sn : RFC6487 4.2.  Serial Number
	if len(cerModel.Sn) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SN is empty",
			Detail: (cerModel.Sn)}
		stateModel.AddError(&stateMsg)
	}
	// max hex is 20, so is 20*2
	// check sn   myssl.c P3797  rescert_serial_number_chk
	if len(cerModel.Sn) > 20*2 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SN is too long",
			Detail: cerModel.Sn}
		stateModel.AddError(&stateMsg)
	}
	isHex, err := regexputil.IsHex(cerModel.Sn)
	if !isHex || err != nil {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SN is not a hexadecimal number",
			Detail: cerModel.Sn}
		stateModel.AddError(&stateMsg)
	}

	// myssl.c P3548  rescert_sig_algs_chk
	// Signature Algorithm  RFC6487 4.3.Signature Algorithm
	if cerModel.SignatureInnerAlgorithm.Name != "sha256WithRSAEncryption" {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Signature Algorithm is not sha256WithRSAEncryption",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// myssl.c P3548  rescert_sig_algs_chk
	if cerModel.SignatureOuterAlgorithm.Name != "sha256WithRSAEncryption" {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Signature Algorithm is not sha256WithRSAEncryption",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// allow len(cerModel.SignatureOuterAlgorithm.Sha256)==0
	if len(cerModel.SignatureOuterAlgorithm.Sha256) > 0 && len(cerModel.SignatureOuterAlgorithm.Sha256) != 767 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "The length of the signature algorithm is wrong",
			Detail: fmt.Sprintf("Signature Outer Sha256 length is %d", len(cerModel.SignatureOuterAlgorithm.Sha256))}
		stateModel.AddError(&stateMsg)
	}
	// Public Key Algorithm  RFC6487 4.3.Signature Algorithm
	// id-dsa id-ecPublicKey id-Ed25519 rsaEncryption
	/*
		x509 PublicKeyAlgorithm:
			RSA
			DSA // Only supported for parsing.
			ECDSA
			Ed25519
	*/
	belogs.Info("validateCerModel(): PublicKeyAlgorithm.Name:", cerModel.PublicKeyAlgorithm.Name,
		"  len(cerModel.PublicKeyAlgorithm.Modulus):", len(cerModel.PublicKeyAlgorithm.Modulus))
	if cerModel.PublicKeyAlgorithm.Name != "rsaEncryption" &&
		cerModel.PublicKeyAlgorithm.Name != "id-dsa" &&
		cerModel.PublicKeyAlgorithm.Name != "id-ecPublicKey" &&
		cerModel.PublicKeyAlgorithm.Name != "id-Ed25519" {
		belogs.Info("validateCerModel(): check PublicKeyAlgorithm.Name fail, cerModel.PublicKeyAlgorithm:", jsonutil.MarshalJson(cerModel.PublicKeyAlgorithm))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "PublicKey Algorithm is not RSA or DSA or ECDSA or Ed25519",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if cerModel.PublicKeyAlgorithm.Name == "rsaEncryption" {
		if len(cerModel.PublicKeyAlgorithm.Modulus) != 512 {
			belogs.Info("validateCerModel(): check PublicKeyAlgorithm.Modulus fail, cerModel.PublicKeyAlgorithm:", jsonutil.MarshalJson(cerModel.PublicKeyAlgorithm),
				"  len(cerModel.PublicKeyAlgorithm.Modulus):", len(cerModel.PublicKeyAlgorithm.Modulus))
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The length of the PublicKey Algorithm’s Modulus is wrong",
				Detail: fmt.Sprintf("PublicKey RSA Modulus length is %d", len(cerModel.PublicKeyAlgorithm.Modulus))}
			stateModel.AddError(&stateMsg)
		}
	}

	if cerModel.PublicKeyAlgorithm.Exponent != 65537 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "PublicKey Algorithm's Exponent is wrong",
			Detail: fmt.Sprintf("PublicKey exponent is %d", cerModel.PublicKeyAlgorithm.Exponent)}
		stateModel.AddError(&stateMsg)
	}

	// issuer    RFC6487 4.4.  Issuer
	if len(cerModel.Issuer) == 0 || len(cerModel.IssuerAll) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Issuer is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// subject RFC6487 4.5. Subject
	if len(cerModel.Subject) == 0 || len(cerModel.SubjectAll) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Subject is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// check subject issuer, sn, from, to, sig
	/* myssl.c P861
		static cf_validator validators[] = {
	    {&cf_get_subject, CF_FIELD_SUBJECT, 1},
	    {&cf_get_issuer, CF_FIELD_ISSUER, 1},
	    {&cf_get_sn, CF_FIELD_SN, 1},
	    {&cf_get_from, CF_FIELD_FROM, 1},
	    {&cf_get_to, CF_FIELD_TO, 1},
	    {&cf_get_sig, CF_FIELD_SIGNATURE, 1},
	  myssl.c P3451  rescert_name_chk
	*/
	if len(cerModel.Issuer) == 0 || len(cerModel.IssuerAll) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Issuer is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	//check time
	// myssl.c P3856 rescert_dates_chk
	// myssl.c P2970-2997
	now := time.Now()
	if cerModel.NotBefore.IsZero() {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotBefore is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if cerModel.NotAfter.IsZero() {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	//thisUpdate precedes nextUpdate.
	if cerModel.NotBefore.After(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotBefore is later than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notBefore is " + convert.Time2StringZone(cerModel.NotBefore)}
		stateModel.AddError(&stateMsg)
	}
	if cerModel.NotAfter.Before(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(cerModel.NotAfter)}
		if conf.Bool("policy::allowStaleCer") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}
	}
	if cerModel.NotBefore.After(cerModel.NotAfter) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter is earlier than NotBefore",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// SubjectAll, IssuerAll: myssl.c  P3450
	//RFC 6487 limits the total number of attributes, not the sequence length explicitly
	// 4.4.  Issuer: /4.5.  Subject: An issuer name MUST contain one instance of the CommonName attribute
	// and MAY contain one instance of the serialNumber attribute.
	split := strings.Split(cerModel.SubjectAll, ",")
	if len(split) > 2 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "There are multiple subjects",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	} else {
		// must contain CN, may contain serialnumber, no other
		foundCN := false
		foundOther := false
		for _, one := range split {
			one = strings.TrimSpace(one)
			if strings.HasPrefix(one, "CN=") {
				foundCN = true
			}
			if !strings.HasPrefix(one, "CN=") && !strings.HasPrefix(one, "SERIALNUMBER=") {
				foundOther = true
			}
		}
		if !foundCN {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Subject attribute MUST contain CN",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if foundOther {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Subject has attribute other than CN and serialnumber",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}
	split = strings.Split(cerModel.IssuerAll, ",")
	if len(split) > 2 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "The Number of issuer attributes should not be greater than 2",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	} else {
		// must contain CN, may contain serialnumber, no other
		foundCN := false
		foundOther := false
		for _, one := range split {
			one = strings.TrimSpace(one)
			belogs.Debug("validateCerModel(): IssuerAll: one:", one)
			if strings.HasPrefix(one, "CN=") {
				foundCN = true
			}
			if !strings.HasPrefix(one, "CN=") && !strings.HasPrefix(one, "SERIALNUMBER=") {
				foundOther = true
			}
		}
		if !foundCN {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Issuer attribute should contain CN",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if foundOther {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Issuer has attribute other than CN and serialnumber",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}

	// check isca and basic_constraints
	// rescert_basic_constraints_chk myssl.c P2002
	if !cerModel.IsCa {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "IsCA must be true",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if !cerModel.BasicConstraintsModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Basic Constraints is not critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if !cerModel.BasicConstraintsModel.BasicConstraintsValid {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "BasicConstraintsValid is not true",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// rescert_ski_chk myssl.c P2130    TODO : is critical ?
	if len(cerModel.Ski) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(cerModel.Ski) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// rescert_aki_chk myssl.c P2247  TODO ? aki.issuer, aki.serial ?
	if !cerModel.IsRoot {
		if len(cerModel.Aki) == 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "AKI is empty",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if len(cerModel.Aki) != 40 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "AKI length is wrong",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}

	//rescert_key_usage_chk myssl.c P2359
	if !cerModel.KeyUsageModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Key Usage is not critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if cerModel.KeyUsageModel.KeyUsage == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Key Usage is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if cerModel.KeyUsageModel.KeyUsageValue != "Certificate Sign, CRL Sign" ||
		cerModel.KeyUsageModel.KeyUsage != 96 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Keyusage is not equal to \"Certificate Sign, CRL Sign\"",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// rfc6487#section-4.8.5   rescert_extended_key_usage_chk, myssl.c P2427
	if len(cerModel.ExtKeyUsages) > 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ExKeyUsage is not empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// chec crldp
	// rescert_crldp_chk myssl.c P2495
	// RFC6487  4.8.6.  CRL Distribution Points
	// TODO no crl extension: subfields
	if cerModel.CrldpModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "CRL Distribution Points are not critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if !cerModel.IsRoot {
		if len(cerModel.CrldpModel.Crldps) == 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "CRL Distribution Points are empty",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if len(cerModel.CrldpModel.Crldps) > 1 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "There are multiple CRL Distribution Points",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
		if len(cerModel.CrldpModel.Crldps) == 1 {
			crldp := cerModel.CrldpModel.Crldps[0]
			u, err := url.Parse(crldp)
			if err != nil {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "CRL Distribution Points are not a legal URL address",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
			}
			if u.Scheme != "rsync" {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "CRL Distribution Points are not an Rsync protocol",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
		}
	} else {
		if len(cerModel.CrldpModel.Crldps) > 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "CRL Distribution Points appear in the root certificate",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}

	// check aia
	// rescert_aia_chk  myssl.c P2645
	if cerModel.AiaModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AIA is critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if !cerModel.IsRoot {
		if len(cerModel.AiaModel.CaIssuers) == 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "AIA is empty",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		} else {
			u, err := url.Parse(cerModel.AiaModel.CaIssuers)
			if err != nil {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "AIA is not a legal URL address",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
			}
			if u.Scheme != "rsync" {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "AIA is not an Rsync protocol",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
		}
	} else {
		if len(cerModel.AiaModel.CaIssuers) > 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "AIA appears in the root certificate",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}

	// check sia
	// rescert_sia_chk myssl.c P2813    RFC6487 4.8.8.1.  SIA for CA Certificates
	if cerModel.SiaModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SIA is critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(cerModel.SiaModel.CaRepository) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "CA Repository is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	} else {
		u, err := url.Parse(cerModel.SiaModel.CaRepository)
		if err != nil {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "CA Repository is not a legal URL address",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
		if u.Scheme != "rsync" {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "CA Repository is not an Rsync protocol",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}
	if len(cerModel.SiaModel.RpkiManifest) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "RpkiMainfest is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	} else {
		u, err := url.Parse(cerModel.SiaModel.RpkiManifest)
		if err != nil {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "RpkiMainfest is not a legal URL address",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
		if u.Scheme != "rsync" {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "RpkiMainfest is not an Rsync protocol",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}
	if len(cerModel.SiaModel.RpkiNotify) > 0 {
		u, err := url.Parse(cerModel.SiaModel.RpkiNotify)
		if err != nil {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "RpkiNotify is not a legal URL address",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "RpkiNotify is not an Http(s) protocol",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}

	// check cert policy CP
	// rescert_cert_policy_chk myssl.c P2958
	if !cerModel.CertPolicyModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Policy must marked critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(cerModel.CertPolicyModel.Cps) == 0 {
		//TODO , not set warning
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Certificate Policies are empty",
			Detail: ""}
		//stateModel.AddWarning(&stateMsg)
		belogs.Debug("validateCerModel(): stateMsg:", stateMsg)
	} else {
		u, err := url.Parse(cerModel.CertPolicyModel.Cps)
		if err != nil {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Certificate Policies are not a legal URL address",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
		if u.Scheme != "https" && u.Scheme != "http" {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Certificate Policies are not Http(s) protocol",
				Detail: ""}
			stateModel.AddError(&stateMsg)
		}
	}

	//check CerIpAddress
	// myssl.c
	//P3277 rescert_ip_asnum_chk
	//P3086 rescert_ip_resources_chk   RFC6487 4.8.10.  IP Resources
	if len(cerModel.CerIpAddressModel.CerIpAddresses) > 0 && !cerModel.CerIpAddressModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "IP address is not critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(cerModel.CerIpAddressModel.CerIpAddresses) > 0 {
		for _, one := range cerModel.CerIpAddressModel.CerIpAddresses {
			if one.AddressFamily != iputil.Ipv4Type && one.AddressFamily != iputil.Ipv6Type {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "IP address is neither IPv4 nor IPv6",
					Detail: ""}
				stateModel.AddError(&stateMsg)
			}
		}
	}
	//P3185  rescert_as_resources_chk   RFC6487  4.8.11.  AS Resources
	if len(cerModel.AsnModel.Asns) > 0 && !cerModel.AsnModel.Critical {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ASN is not critical",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// check critical
	// myssl.c P3371
	/*
			static int supported_nids[] = {
		        NID_key_usage,          // 83
		        NID_basic_constraints,  // 87
		        NID_certificate_policies,       // 89
		        NID_sbgp_ipAddrBlock,   // 290
		        NID_sbgp_autonomousSysNum,      // 291
		    };
	*/

	//  check  unique ID ??TODO   myssl.c P3878 rescert_subj_iss_UID_chk no exist issuerUniqueID and subjectUniqueID

	//  myssl.c P3901  rescert_extensions_chk
	for _, ext := range cerModel.ExtensionModels {
		belogs.Debug("validateCerModel():ext:", cerModel.FilePath, cerModel.FileName, jsonutil.MarshalJson(ext))

		/*https://datatracker.ietf.org/doc/html/rfc6487#section-4.8
		{"oid":"2.5.29.32","critical":true,"name":"CertPolicy"} //Certificate Policies:MUST be present /  critical
		{"oid":"2.5.29.19","critical":true,"name":"BasicConstraints"}  //Basic Constraints:  MUST be present/  critical
		{"oid":"2.5.29.15","critical":true,"name":"KeyUsage"} //Key Usage: MUST be present/  critical
		{"oid":"1.3.6.1.5.5.7.1.8","critical":true,"name":"Asn"}  //AS Resources: MUST be present / critical
		{"oid":"1.3.6.1.5.5.7.1.7","critical":true,"name":"CerIpAddress"} //IP Resources: MUST be present/critical

		{"oid":"2.5.29.14","critical":false,"name":"Ski"} //Subject Key Identifier: MUST appear / non-critical
		{"oid":"1.3.6.1.5.5.7.1.11","critical":false,"name":"Sia"} //Subject Information Access: MUST be present / non-critical

		{"oid":"1.3.6.1.5.5.7.1.1","critical":false,"name":"Aia"} //Authority Information Access:  MUST be present or self-signed/ non-critical
		{"oid":"2.5.29.35","critical":false,"name":"Aki"} //Authority Key Identifier:MUST appear or self-signed/ non-critical
		{"oid":"2.5.29.31","critical":false,"name":"Crldp"} //CRL Distribution Points: MUST be present or self-signed  / non-critical
		// Extended Key Usage: MUST NOT appear no-critical
		*/
		// shoud not present in cer, check by cermodel.CerExtensionOids
		if len(ext.Name) == 0 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "This extension name is empty",
				Detail: ext.Oid}
			stateModel.AddError(&stateMsg)
			continue
		}

		// check Critical value
		switch ext.Name {
		case "CertPolicy":
			fallthrough
		case "BasicConstraints":
			fallthrough
		case "KeyUsage":
			fallthrough
		case "Asn":
			fallthrough
		case "CerIpAddress":
			if !ext.Critical {
				belogs.Error("validateCerModel():ext should Critical, but now is non-Critical:", cerModel.FilePath, cerModel.FileName,
					jsonutil.MarshalJson(ext))
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "This Extension must be critical",
					Detail: ext.Name}
				stateModel.AddError(&stateMsg)
			}

		case "Ski":
			fallthrough
		case "Sia":
			if ext.Critical {
				belogs.Error("validateCerModel():ext should not Critical, but now is Critical:", cerModel.FilePath, cerModel.FileName,
					jsonutil.MarshalJson(ext))
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "This Extension must be none-critical",
					Detail: ext.Name}
				stateModel.AddError(&stateMsg)
			}

		case "Aia":
			fallthrough
		case "Aki":
			fallthrough
		case "Crldp":
			if !cerModel.IsRoot && ext.Critical {
				belogs.Error("validateCerModel():cer is not root, and ext should not Critical, but now is Critical:", cerModel.FilePath, cerModel.FileName,
					jsonutil.MarshalJson(ext), " is root:", cerModel.IsRoot)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "This Extension must be none-critical",
					Detail: jsonutil.MarshalJson(ext)}
				stateModel.AddError(&stateMsg)
			}

		default:
			belogs.Error("validateCerModel():in default,ext should not present:", cerModel.FilePath, cerModel.FileName,
				jsonutil.MarshalJson(ext), " is root:", cerModel.IsRoot)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "This extension should not be present",
				Detail: jsonutil.MarshalJson(ext)}
			stateModel.AddError(&stateMsg)
		}

	}
	belogs.Debug("validateCerModel():filePath, fileName:", cerModel.FilePath, cerModel.FileName,
		"  stateModel:", jsonutil.MarshalJson(stateModel))
	return nil
}

func parseCerSimpleModel(fileModel *model.FileModel) (parseCerSimple model.ParseCerSimple, err error) {
	belogs.Debug("parseCerSimpleModel():fileModel:", jsonutil.MarshalJson(fileModel))

	parseCerSimple, err = asn1.ParseCerSimpleModelByAsn1(fileModel)
	if err == nil {
		belogs.Debug("parseCerSimpleModel():ParseCerSimpleModelByAsn1 ok, parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))
		return parseCerSimple, nil
	}
	belogs.Error("parseCerSimpleModel(): ParseCerSimpleModelByAsn1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)

	parseCerSimple, err = openssl.ParseCerSimpleModelByOpenssl(fileModel)
	if err != nil {
		belogs.Error("parseCerSimpleModel(): ParseCerSimpleModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return parseCerSimple, err
	}
	belogs.Debug("parseCerSimpleModel():ParseCerSimpleModelByOpenssl ok, parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))
	return parseCerSimple, nil
}
