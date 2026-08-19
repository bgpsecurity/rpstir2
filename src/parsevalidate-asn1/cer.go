package parsevalidateasn1

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/asn1util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

type AddressFamilyBlockAsn1 struct {
	AddressFamilyAsn1 []byte
	AddressChoiceAsn1 asn1.RawValue
}
type AddressRangeAsn1 struct {
	MinAddressAsn1 asn1.BitString
	MaxAddressAsn1 asn1.BitString
}

func parseCerIpAddressModelByAsn1(data []byte) (cerIpAddressModel model.CerIpAddressModel, noCerIpAddress bool, err error) {
	start := time.Now()
	belogs.Debug("parseCerIpAddressModelByAsn1(): len(data):", len(data))
	var addressFamilyBlockAsn1s []AddressFamilyBlockAsn1
	_, err = asn1.Unmarshal(data, &addressFamilyBlockAsn1s)
	if err != nil {
		belogs.Error("parseCerIpAddressModelByAsn1(): Unmarshal data fail, len(data):", len(data), err)
		return cerIpAddressModel, false, err
	}
	noCerIpAddress = true
	belogs.Debug("parseCerIpAddressModelByAsn1(): len(addressFamilyBlockAsn1s):", len(addressFamilyBlockAsn1s),
		"   addressFamilyBlockAsn1s:", jsonutil.MarshalJson(addressFamilyBlockAsn1s))
	for i := range addressFamilyBlockAsn1s {

		// compatible
		var ipType uint64
		if len(addressFamilyBlockAsn1s[i].AddressFamilyAsn1) == 2 {
			if addressFamilyBlockAsn1s[i].AddressFamilyAsn1[1] == 1 {
				ipType = 1
			} else if addressFamilyBlockAsn1s[i].AddressFamilyAsn1[1] == 2 {
				ipType = 2
			}
		} else if len(addressFamilyBlockAsn1s[i].AddressFamilyAsn1) == 1 {
			if addressFamilyBlockAsn1s[i].AddressFamilyAsn1[0] == 1 {
				ipType = 1
			} else if addressFamilyBlockAsn1s[i].AddressFamilyAsn1[0] == 2 {
				ipType = 2
			}
		}
		if ipType == 0 {
			belogs.Error("parseCerIpAddressModelByAsn1(): ipType==0, fail:", ipType)
			return cerIpAddressModel, false, errors.New("ipType is error")
		}
		belogs.Debug("parseCerIpAddressModelByAsn1(): addressFamilyBlockAsn1s[i]:", addressFamilyBlockAsn1s[i],
			"   ipType:", ipType)

		if addressFamilyBlockAsn1s[i].AddressChoiceAsn1.Tag == asn1.TagNull {
			// is null
			belogs.Debug("parseCerIpAddressModelByAsn1():AddressChoiceAsn1 is TagNull:", addressFamilyBlockAsn1s[i].AddressChoiceAsn1.Tag)

		} else if addressFamilyBlockAsn1s[i].AddressChoiceAsn1.Tag == asn1.TagSequence {
			noCerIpAddress = false
			// have ips
			belogs.Debug("parseCerIpAddressModelByAsn1():addressFamilyBlockAsn1s[i] is TagSequence:", addressFamilyBlockAsn1s[i].AddressChoiceAsn1.Tag)

			var addressChoiceAsn1s []asn1.RawValue
			_, err = asn1.Unmarshal(addressFamilyBlockAsn1s[i].AddressChoiceAsn1.FullBytes, &addressChoiceAsn1s)
			if err != nil {
				belogs.Error("parseCerIpAddressModelByAsn1():addressChoiceAsn1s Unmarshal fail:",
					convert.PrintBytesOneLine(addressFamilyBlockAsn1s[i].AddressChoiceAsn1.FullBytes),
					err)
				return cerIpAddressModel, false, err
			}
			belogs.Debug("parseCerIpAddressModelByAsn1(): len(addressChoiceAsn1s):", len(addressChoiceAsn1s))

			cerIpAddress := model.CerIpAddress{}
			cerIpAddress.AddressFamily = uint64(ipType)
			for j := range addressChoiceAsn1s {
				if addressChoiceAsn1s[j].Tag == asn1.TagBitString {
					addressPrefix, err := ParseBytesToAddressPrefix(addressChoiceAsn1s[j].Bytes, int(ipType))
					if err != nil {
						// try again
						addressPrefix, err = ParseBytesToAddressPrefix(addressChoiceAsn1s[j].FullBytes, int(ipType))
						if err != nil {
							belogs.Error("parseCerIpAddressModelByAsn1():TagBitString ParseBytesToAddressPrefix fail:",
								convert.PrintBytesOneLine(addressChoiceAsn1s[j].Bytes), "  ipType:", ipType, err)
							return cerIpAddressModel, false, err
						}
					}
					cerIpAddress.AddressPrefix, err = iputil.TrimAddressPrefixZero(addressPrefix, int(ipType))
					if err != nil {
						belogs.Error("parseCerIpAddressModelByAsn1():TrimAddressPrefixZero fail, addressPrefix:", addressPrefix, "  ipType:", ipType, err)
						return cerIpAddressModel, false, err
					}
					cerIpAddress.RangeStart, cerIpAddress.RangeEnd, err = iputil.AddressPrefixToHexRange(addressPrefix, int(ipType))
					if err != nil {
						belogs.Error("parseCerIpAddressModelByAsn1():AddressPrefixToHexRange fail, addressPrefix:", addressPrefix, "  ipType:", ipType, err)
						return cerIpAddressModel, false, err
					}
					cerIpAddress.AddressPrefixRange = jsonutil.MarshalJson(addressPrefix)
					belogs.Debug("parseCerIpAddressModelByAsn1(): AddressPrefix cerIpAddress:", jsonutil.MarshalJson(cerIpAddress))

				} else if addressChoiceAsn1s[j].Tag == asn1.TagSequence {

					cerIpAddress.Min, cerIpAddress.Max, err = ParseToAddressMinMax(addressChoiceAsn1s[j].FullBytes, int(ipType))
					if err != nil {
						belogs.Error("parseCerIpAddressModelByAsn1(): ParseToAddressMinMax fail:",
							convert.PrintBytesOneLine(addressChoiceAsn1s[j].FullBytes), "  ipType:", ipType, err)
						return cerIpAddressModel, false, err
					}
					cerIpAddress.RangeStart, err = iputil.IpStrToHexString(cerIpAddress.Min, int(ipType))
					if err != nil {
						belogs.Error("parseCerIpAddressModelByAsn1():IpNetToHexString min err:", err)
						return cerIpAddressModel, false, err
					}
					cerIpAddress.RangeEnd, err = iputil.IpStrToHexString(cerIpAddress.Max, int(ipType))
					if err != nil {
						belogs.Error("parseCerIpAddressModelByAsn1():IpNetToHexString max err:", err)
						return cerIpAddressModel, false, err
					}
					addressPrefixRanges := iputil.IpRangeToAddressPrefixRanges(cerIpAddress.Min, cerIpAddress.Max)
					cerIpAddress.AddressPrefixRange = jsonutil.MarshalJson(addressPrefixRanges)
					belogs.Debug("parseCerIpAddressModelByAsn1(): minMax cerIpAddress:", jsonutil.MarshalJson(cerIpAddress))
				}
				cerIpAddressModel.CerIpAddresses = append(cerIpAddressModel.CerIpAddresses, cerIpAddress)
			}
		}
	}
	belogs.Debug("parseCerIpAddressModelByAsn1(): cerIpAddressModel:", jsonutil.MarshalJson(cerIpAddressModel), "   time(s):", time.Since(start))
	return cerIpAddressModel, noCerIpAddress, nil
}

type AsnBlockAsn1 struct {
	AsnChoiceAsn1 AsnChoiceAsn1 `asn1:"optional,tag:0"`
}

type AsnChoiceAsn1 struct {
	AsnSequenceAsn1 []asn1.RawValue `asn1:"optional"`
}
type AnsRangeAsn1 struct {
	MinAsn1 int
	MaxAsn1 int
}

func parseAsnModelByAsn1(data []byte) (asnModel model.AsnModel, noAsn bool, err error) {
	start := time.Now()
	belogs.Debug("parseAsnModelByAsn1(): len(data):", len(data))
	var asnBlockAsn1 AsnBlockAsn1
	_, err = asn1.Unmarshal(data, &asnBlockAsn1)
	if err != nil {
		belogs.Error("parseAsnModelByAsn1(): Unmarshal data fail, data:", hex.EncodeToString(data), "  len(data):", len(data), err)
		return asnModel, false, err
	}
	asnModel.Critical = true
	noAsn = true
	belogs.Debug("parseAsnModelByAsn1(): asnBlockAsn1:", jsonutil.MarshalJson(asnBlockAsn1))

	for i := range asnBlockAsn1.AsnChoiceAsn1.AsnSequenceAsn1 {
		if asnBlockAsn1.AsnChoiceAsn1.AsnSequenceAsn1[i].Tag == asn1.TagInteger {
			noAsn = false
			asn := model.NewAsn()
			b := asnBlockAsn1.AsnChoiceAsn1.AsnSequenceAsn1[i].Bytes
			asnInt := big.NewInt(0).SetBytes(b)
			asn.Asn = asnInt.Int64()
			belogs.Debug("parseAsnModelByAsn1(): TagInteger asn:", jsonutil.MarshalJson(asn))
			asnModel.Asns = append(asnModel.Asns, asn)
		} else if asnBlockAsn1.AsnChoiceAsn1.AsnSequenceAsn1[i].Tag == asn1.TagSequence {
			noAsn = false
			asn := model.NewAsn()
			var ansRangeAsn1 AnsRangeAsn1
			_, err := asn1.Unmarshal(asnBlockAsn1.AsnChoiceAsn1.AsnSequenceAsn1[i].FullBytes, &ansRangeAsn1)
			if err != nil {
				belogs.Error("parseAsnModelByAsn1(): Unmarshal AsnSequenceAsn1 fail, len(asnBlockAsn1.AsnSequenceAsn1.AsnSequenceAsn1[i].FullBytes):",
					len(asnBlockAsn1.AsnChoiceAsn1.AsnSequenceAsn1[i].FullBytes), err)
				continue
			}
			asn.Min = int64(ansRangeAsn1.MinAsn1)
			asn.Max = int64(ansRangeAsn1.MaxAsn1)
			belogs.Debug("parseAsnModelByAsn1(): TagSequence asn:", jsonutil.MarshalJson(asn))
			asnModel.Asns = append(asnModel.Asns, asn)
		}
	}
	belogs.Debug("parseAsnModelByAsn1(): asnModel:", jsonutil.MarshalJson(asnModel), "   time(s):", time.Since(start))
	return asnModel, noAsn, nil
}

type InfoAccessAsn1 struct {
	InfoAccessOid      asn1.ObjectIdentifier
	InfoAccessByteAsn1 []byte `asn1:"implicit,tag:6"`
	//Value string `asn1:"implicit,tag:6"`
}

func parseInfoAccessAsn1ByAsn1(data []byte) (infoAccessAsn1s []InfoAccessAsn1, err error) {
	belogs.Debug("parseInfoAccessAsn1ByAsn1(): len(data):", len(data))
	infoAccessAsn1s = make([]InfoAccessAsn1, 0)
	_, err = asn1.Unmarshal(data, &infoAccessAsn1s)
	if err != nil {
		belogs.Error("parseInfoAccessAsn1ByAsn1(): Unmarshal data fail, len(data):", len(data), err)
		return nil, err
	}
	belogs.Debug("parseInfoAccessAsn1ByAsn1(): infoAccessAsn1s:", jsonutil.MarshalJson(infoAccessAsn1s))
	return infoAccessAsn1s, nil
}

func parseAiaModelByAsn1(data []byte) (aiaModel model.AiaModel, err error) {
	belogs.Debug("parseAiaModelByAsn1(): len(data):", len(data))
	infoAccessAsn1s, err := parseInfoAccessAsn1ByAsn1(data)
	if err != nil {
		belogs.Error("parseAiaModelByAsn1(): parseInfoAccessAsn1ByAsn1 fail, len(data):", len(data), err)
		return aiaModel, err
	}
	for i := range infoAccessAsn1s {
		if infoAccessAsn1s[i].InfoAccessOid.String() == "1.3.6.1.5.5.7.48.2" {
			aiaModel.CaIssuers = strings.TrimSpace(string(infoAccessAsn1s[i].InfoAccessByteAsn1))
		}
	}
	belogs.Debug("parseAiaModelByAsn1(): aiaModel.CaIssuers:", aiaModel.CaIssuers)
	return aiaModel, nil
}

func parseSiaModelByAsn1(data []byte) (siaModel model.SiaModel, err error) {
	infoAccessAsn1s, err := parseInfoAccessAsn1ByAsn1(data)
	if err != nil {
		belogs.Error("parseSiaModelByAsn1(): parseInfoAccessAsn1ByAsn1 fail, len(data):", len(data), err)
		return siaModel, err
	}
	for i := range infoAccessAsn1s {
		if infoAccessAsn1s[i].InfoAccessOid.String() == "1.3.6.1.5.5.7.48.5" {
			siaModel.CaRepository = strings.TrimSpace(string(infoAccessAsn1s[i].InfoAccessByteAsn1))
		} else if infoAccessAsn1s[i].InfoAccessOid.String() == "1.3.6.1.5.5.7.48.10" {
			siaModel.RpkiManifest = strings.TrimSpace(string(infoAccessAsn1s[i].InfoAccessByteAsn1))
		} else if infoAccessAsn1s[i].InfoAccessOid.String() == "1.3.6.1.5.5.7.48.13" {
			siaModel.RpkiNotify = strings.TrimSpace(string(infoAccessAsn1s[i].InfoAccessByteAsn1))
		} else if infoAccessAsn1s[i].InfoAccessOid.String() == "1.3.6.1.5.5.7.48.11" {
			siaModel.SignedObject = strings.TrimSpace(string(infoAccessAsn1s[i].InfoAccessByteAsn1))
		}
	}
	belogs.Debug("parseSiaModelByAsn1(): siaModel:", jsonutil.MarshalJson(siaModel))
	return siaModel, nil
}

type PolicyInformation struct {
	PolicyIdentifier asn1.ObjectIdentifier
	PolicyQualifiers []PolicyQualifierInfo `asn1:"optional"`
}
type PolicyQualifierInfo struct {
	PolicyQualifierId asn1.ObjectIdentifier
	Qualifier         string `asn1:"ia5"` // 或使用 RawValue 判断
}

type PolicyBlockAsn1 struct {
	PolicyOidAsn1 asn1.ObjectIdentifier
	PolicyAsn1s   []PolicyAsn1 `asn1:"optional"`
}

type PolicyAsn1 struct {
	PolicyOidAsn1 asn1.ObjectIdentifier
	PolicyUrl     string
}

func parseCertPolicyModelByAsn1(data []byte) (certPolicyModel model.CertPolicyModel, err error) {
	belogs.Debug("parseCertPolicyModelByAsn1(): len(data):", len(data), " data:", convert.PrintBytesOneLine(data))

	var policyInfos []PolicyInformation
	_, err = asn1.Unmarshal(data, &policyInfos)
	if err == nil {
		belogs.Debug("parseCertPolicyModelByAsn1(): len(policyInfos):", len(policyInfos))

		// 提取第一个 CPS URI（id-qt-cps = 1.3.6.1.5.5.7.2.1）
		const cpsOid = "1.3.6.1.5.5.7.2.1"
		for _, info := range policyInfos {
			for _, qualifier := range info.PolicyQualifiers {
				if qualifier.PolicyQualifierId.String() == cpsOid {
					certPolicyModel.Cps = qualifier.Qualifier
					belogs.Debug("parseCertPolicyModelByAsn1(): found CPS:", certPolicyModel.Cps)
					return certPolicyModel, nil // 找到第一个即返回
				}
			}
		}
		belogs.Debug("parseCertPolicyModelByAsn1(): no CPS found, certPolicyModel:", jsonutil.MarshalJson(certPolicyModel))
		return certPolicyModel, nil
	} else {
		// try again
		policyBlockAsn1s := make([]PolicyBlockAsn1, 0)
		_, err = asn1.Unmarshal(data, &policyBlockAsn1s)
		if err != nil {
			belogs.Error("parseCertPolicyModelByAsn1(): Unmarshal fail, len(data):", len(data),
				"  data:", hex.EncodeToString(data), err)
			return certPolicyModel, err
		}
		belogs.Debug("parseCertPolicyModelByAsn1(): len(policyBlockAsn1s):", len(policyBlockAsn1s))
		for i := range policyBlockAsn1s {
			policyAsn1 := policyBlockAsn1s[i].PolicyAsn1s
			for j := range policyAsn1 {
				certPolicyModel.Cps = policyAsn1[j].PolicyUrl
				break
			}
		}
		belogs.Debug("parseCertPolicyModelByAsn1(): certPolicyModel:", jsonutil.MarshalJson(certPolicyModel))
		return certPolicyModel, nil
	}
}

func parseSignatureInnerAlgorithmByAsn1(signatureAlgorithm x509.SignatureAlgorithm) (signatureInnerAlgorithm model.Sha256RsaModel, err error) {
	belogs.Debug("parseSignatureInnerAlgorithmByAsn1(): signatureAlgorithm:", jsonutil.MarshalJson(signatureAlgorithm))
	if signatureAlgorithm == x509.SHA256WithRSA {
		signatureInnerAlgorithm.Name = "sha256WithRSAEncryption"
	}
	return signatureInnerAlgorithm, nil
}
func parseSignatureOuterAlgorithmByAsn1(data []byte) (signatureOuterAlgorithm model.Sha256RsaModel, err error) {
	signatureOuterAlgorithm.Name = "sha256WithRSAEncryption"
	/* if want get signatureOuterAlgorithm.Sha256, should asn1.Unmarshal(data,&Certificate),
	   but it is too cumbersome and has little impact on RPKI parsing validation, so just ignore

	type Certificate struct {
		//Raw                asn1.RawContent
		TBSCertificate     TbsCertificate
		SignatureAlgorithm AlgorithmIdentifier
		SignatureValue     BitString
	}

	*/
	return
}

/*
	type PublicKey struct {
		N *big.Int // modulus
		E int      // public exponent
	}
*/
type RsaNEModel struct {
	N *big.Int `json:"N"`
	E *big.Int `json:"E"` // E int
}

func parsePublicKeyAlgorithmByAsn1(publicKeyAlgorithmx509 x509.PublicKeyAlgorithm,
	publicKey any) (publicKeyAlgorithmModel model.RsaModel, err error) {
	publicKeyJson := jsonutil.MarshalJson(publicKey)
	belogs.Debug("parsePublicKeyAlgorithmByAsn1(): publicKeyAlgorithmx509:", jsonutil.MarshalJson(publicKeyAlgorithmx509),
		"  publicKeyJson:", publicKeyJson)
	/*
		x509 PublicKeyAlgorithm:
			RSA
			DSA // Only supported for parsing.
			ECDSA
			Ed25519
	*/
	if publicKeyAlgorithmx509 == x509.RSA {
		publicKeyAlgorithmModel.Name = "rsaEncryption"
		rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
		if ok {
			publicKeyAlgorithmModel.Exponent = uint64(rsaPublicKey.E)
			publicKeyAlgorithmModel.Modulus = fmt.Sprintf("%x", rsaPublicKey.N)
		} else {
			// try again
			var rsaNEModel RsaNEModel
			err = jsonutil.UnmarshalJson(publicKeyJson, &rsaNEModel)
			if err != nil {
				belogs.Error("parsePublicKeyAlgorithmByAsn1(): Unmarshal rsaNEModel fail, publicKeyJson:", publicKeyJson, err)
				return publicKeyAlgorithmModel, err
			}
			if rsaNEModel.E != nil || rsaNEModel.N != nil {
				publicKeyAlgorithmModel.Exponent = uint64(rsaNEModel.E.Int64())
				publicKeyAlgorithmModel.Modulus = fmt.Sprintf("%x", rsaNEModel.N)
			} else {
				belogs.Error("parsePublicKeyAlgorithmByAsn1():rsaNEModel.E or N is nil:", jsonutil.MarshalJson(rsaNEModel))
				return publicKeyAlgorithmModel, errors.New("fail to parse rsaNEModel")
			}
		}
	} else if publicKeyAlgorithmx509 == x509.DSA {
		publicKeyAlgorithmModel.Name = "id-dsa"
	} else if publicKeyAlgorithmx509 == x509.ECDSA {
		publicKeyAlgorithmModel.Name = "id-ecPublicKey"
	} else if publicKeyAlgorithmx509 == x509.Ed25519 {
		publicKeyAlgorithmModel.Name = "id-Ed25519"
	}
	belogs.Debug("parsePublicKeyAlgorithmByAsn1(): publicKeyAlgorithmModel:", jsonutil.MarshalJson(publicKeyAlgorithmModel))
	return
}

func ParseCerModelByAsn1(fileModel *model.FileModel, isForEeCer bool,
	cerModel *model.CerModel, stateModel *model.StateModel) (err error) {
	belogs.Debug("ParseCerModelByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel), "  isForEeCer:", isForEeCer)
	fileByte, err := os.ReadFile(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCerModelByAsn1(): ReadFile fail, fileModel.TempFilePathName:", fileModel.TempFilePathName, err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	return parseCerModelUseFileByteByAsn1(fileModel, fileByte, isForEeCer, cerModel, stateModel)
}

func parseCerModelUseFileByteByAsn1(fileModel *model.FileModel, fileByte []byte, isForEeCer bool,
	cerModel *model.CerModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("parseCerModelUseFileByteByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel), "   len(fileByte):", len(fileByte),
		"  isForEeCer:", isForEeCer)
	cer, err := x509.ParseCertificate(fileByte)
	if err != nil {
		belogs.Error("parseCerModelUseFileByteByAsn1():ParseCertificate fail: len(fileByte):", len(fileByte),
			"   fileModel:", jsonutil.MarshalJson(fileModel), err,
			"  time(s):", time.Since(start))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by asn1",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	cerModel.Sn = fmt.Sprintf("%x", cer.SerialNumber)
	cerModel.Version = cer.Version
	cerModel.BasicConstraintsModel.BasicConstraintsValid = cer.BasicConstraintsValid
	cerModel.NotBefore = cer.NotBefore.Local()
	cerModel.NotAfter = cer.NotAfter.Local()
	cerModel.Subject = cer.Subject.CommonName
	cerModel.SubjectAll = cer.Subject.String()
	cerModel.Issuer = cer.Issuer.CommonName
	cerModel.IssuerAll = cer.Issuer.String()
	cerModel.KeyUsageModel.KeyUsage = int(cer.KeyUsage)
	cerModel.ExtKeyUsages = asn1util.ExtKeyUsagesToInts(cer.ExtKeyUsage)
	cerModel.IsCa = cer.IsCA
	cerModel.SignatureInnerAlgorithm, _ = parseSignatureInnerAlgorithmByAsn1(cer.SignatureAlgorithm)
	cerModel.SignatureOuterAlgorithm, _ = parseSignatureOuterAlgorithmByAsn1(fileByte)
	cerModel.PublicKeyAlgorithm, err = parsePublicKeyAlgorithmByAsn1(cer.PublicKeyAlgorithm, cer.PublicKey)
	if err != nil {
		belogs.Error("parseCerModelUseFileByteByAsn1(): parsePublicKeyAlgorithmByAsn1 fail, cer.PublicKeyAlgorithm:", jsonutil.MarshalJson(cer.PublicKeyAlgorithm),
			"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		// no return err
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse publicKey Algorithm by asn1",
			Detail: err.Error()}
		stateModel.AddWarning(&stateMsg)
	}
	cerModel.SubjectPublicKeyInfo = cer.RawSubjectPublicKeyInfo

	//SKI
	cerModel.Ski = convert.Bytes2String(cer.SubjectKeyId)
	//AKI
	cerModel.Aki = convert.Bytes2String(cer.AuthorityKeyId)
	if cerModel.Ski == cerModel.Aki || len(cerModel.Aki) == 0 {
		cerModel.IsRoot = true
	} else {
		cerModel.IsRoot = false
	}

	//CRLDPS
	cerModel.CrldpModel.Crldps = make([]string, 0)
	for _, crldp := range cer.CRLDistributionPoints {
		cerModel.CrldpModel.Crldps = append(cerModel.CrldpModel.Crldps, crldp)
	}

	var noCerIpAddress, noAsn bool
	cerModel.ExtensionModels = make([]model.ExtensionModel, 0)
	belogs.Debug("parseCerModelUseFileByteByAsn1(): len(cer.Extensions):", len(cer.Extensions))
	for _, ext := range cer.Extensions {
		extensionModel := model.ExtensionModel{
			Oid:      ext.Id.String(),
			Critical: ext.Critical,
		}
		if name, ok := model.CerExtensionOids[ext.Id.String()]; ok {
			extensionModel.Name = name
		}
		cerModel.ExtensionModels = append(cerModel.ExtensionModels, extensionModel)

		if extensionModel.Oid == "1.3.6.1.5.5.7.1.7" {
			// IpBlocks
			cerModel.CerIpAddressModel, noCerIpAddress, err = parseCerIpAddressModelByAsn1(ext.Value)
			if err != nil {
				belogs.Error("parseCerModelUseFileByteByAsn1(): parseCerIpAddressModelByAsn1 fail, len(ext.Value):", len(ext.Value),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				// no return err
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ipaddress by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			cerModel.CerIpAddressModel.Critical = ext.Critical
			belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.CerIpAddressModel:", jsonutil.MarshalJson(cerModel.CerIpAddressModel), "  noCerIpAddress:", noCerIpAddress)
		} else if extensionModel.Oid == "1.3.6.1.5.5.7.1.8" {
			if !isForEeCer {
				// Asns
				cerModel.AsnModel, noAsn, err = parseAsnModelByAsn1(ext.Value)
				if err != nil {
					belogs.Error("parseCerModelUseFileByteByAsn1(): parseAsnModelByAsn1 fail, len(ext.Value):", len(ext.Value),
						"   fileModel:", jsonutil.MarshalJson(fileModel), err)
					// no return err
					stateMsg := model.StateMsg{Stage: "parsevalidate",
						Fail:   "Fail to parse asn by asn1",
						Detail: err.Error()}
					stateModel.AddError(&stateMsg)
					continue
				}
				belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.AsnModel:", jsonutil.MarshalJson(cerModel.AsnModel), "  noAsn:", noAsn)
			}
		} else if extensionModel.Oid == "1.3.6.1.5.5.7.1.1" {
			// authorityInfoAccess cerModel.AiaModel,
			cerModel.AiaModel, err = parseAiaModelByAsn1(ext.Value)
			if err != nil {
				belogs.Error("parseCerModelUseFileByteByAsn1(): parseAiaModelByAsn1 fail, len(ext.Value):", len(ext.Value),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				// no return err
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse aia by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			cerModel.AiaModel.Critical = ext.Critical
			belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.AiaModel:", jsonutil.MarshalJson(cerModel.AiaModel))
		} else if extensionModel.Oid == "1.3.6.1.5.5.7.1.11" {
			// cerModel.SiaModel
			cerModel.SiaModel, err = parseSiaModelByAsn1(ext.Value)
			if err != nil {
				belogs.Error("parseCerModelUseFileByteByAsn1(): parseSiaModelByAsn1 fail, len(ext.Value):", len(ext.Value),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				// no return err
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse sia by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			cerModel.SiaModel.Critical = ext.Critical
			belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.SiaModel:", jsonutil.MarshalJson(cerModel.SiaModel))
		} else if extensionModel.Oid == "2.5.29.15" {
			// already have KeyUsageModel.KeyUsage
			cerModel.KeyUsageModel.Critical = ext.Critical
			cerModel.KeyUsageModel.KeyUsageValue = "Certificate Sign, CRL Sign"
			belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.KeyUsageModel:", jsonutil.MarshalJson(cerModel.KeyUsageModel))

		} else if extensionModel.Oid == "2.5.29.31" {
			// already have cerModel.CrldpModel.Crldps
			cerModel.CrldpModel.Critical = ext.Critical
			belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.CrldpModel:", jsonutil.MarshalJson(cerModel.CrldpModel))
		} else if extensionModel.Oid == "2.5.29.32" {
			// cerModel.CertPolicyModel, EE has no policy
			if !isForEeCer {
				cerModel.CertPolicyModel, err = parseCertPolicyModelByAsn1(ext.Value)
				if err != nil {
					belogs.Error("parseCerModelUseFileByteByAsn1(): parseCertPolicyModelByAsn1 fail, len(ext.Value):", len(ext.Value),
						"   fileModel:", jsonutil.MarshalJson(fileModel), err)
					// no return err
					stateMsg := model.StateMsg{Stage: "parsevalidate",
						Fail:   "Fail to parse policy by asn1",
						Detail: err.Error()}
					stateModel.AddError(&stateMsg)
					continue
				}
				cerModel.CertPolicyModel.Critical = ext.Critical
				belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.CertPolicyModel:", jsonutil.MarshalJson(cerModel.CertPolicyModel))
			}
		} else if extensionModel.Oid == "2.5.29.19" {
			// already have BasicConstraintsModel.BasicConstraintsValid
			cerModel.BasicConstraintsModel.Critical = ext.Critical
			belogs.Debug("parseCerModelUseFileByteByAsn1(): cerModel.BasicConstraintsModel:", jsonutil.MarshalJson(cerModel.BasicConstraintsModel))
		} else {
			belogs.Debug("parseCerModelUseFileByteByAsn1(): not parse OID:", extensionModel.Oid)
		}
	}

	belogs.Debug("parseCerModelUseFileByteByAsn1():cerModel:", cerModel.String(), "  time(s):", time.Since(start))
	return nil
}

func ParseCerSimpleModelByAsn1(fileModel *model.FileModel) (parseCerSimple model.ParseCerSimple, err error) {
	start := time.Now()
	belogs.Debug("ParseCerSimpleModelByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel))
	var cerModel model.CerModel
	var stateModel model.StateModel
	err = ParseCerModelByAsn1(fileModel, true, &cerModel, &stateModel)
	if err != nil {
		belogs.Error("ParseCerSimpleModelByAsn1(): ParseCerModelByAsn1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return
	}
	parseCerSimple.RpkiNotify = cerModel.SiaModel.RpkiNotify
	parseCerSimple.CaRepository = cerModel.SiaModel.CaRepository
	parseCerSimple.SubjectPublicKeyInfo = cerModel.SubjectPublicKeyInfo
	belogs.Debug("ParseCerSimpleModelByAsn1():parseCerSimple:", jsonutil.MarshalJson(parseCerSimple), "  time(s):", time.Since(start))
	return
}
