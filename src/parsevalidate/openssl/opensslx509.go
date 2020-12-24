package openssl

import (
	"bytes"
	"strconv"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	iputil "github.com/cpusoft/goutil/iputil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
)

func ParseCerIpAddressModelByOpensslResults(results []string) (cerIpAddressModel model.CerIpAddressModel, noCerIpAddress bool, err error) {
	cerIpAddressModel.CerIpAddresses = make([]model.CerIpAddress, 0)
	start := -1
	for i, one := range results {
		if strings.Contains(one, "sbgp-ipAddrBlock:") {
			if strings.Contains(one, "critical") {
				cerIpAddressModel.Critical = true
			} else {
				cerIpAddressModel.Critical = false
			}
			start = i + 1
			break
		}
	}
	if start < 0 {
		belogs.Debug("ParseCerIpAddressModelByOpensslResults(): no noCerIpAddress found")
		return cerIpAddressModel, true, nil
	}
	ips := results[start:]
	var ipType int
	for _, ipTmp := range ips {
		ipp := strings.TrimSpace(ipTmp)
		belogs.Debug("ParseCerIpAddressModelByOpensslResults():ipp:", ipp)

		if strings.Contains(ipp, "IPv4:") {
			//ignore
			ipType = iputil.Ipv4Type
			continue
		}
		if strings.Contains(ipp, "IPv6:") {
			//ignore
			ipType = iputil.Ipv6Type
			continue
		}
		if len(ipp) == 0 {
			// end
			break
		}

		cerIpAddress := model.CerIpAddress{}
		cerIpAddress.AddressFamily = uint64(ipType)
		split := strings.Split(ipp, "-")
		if len(split) == 2 {
			//ipAddressOrRange.AddressRange := certstruct.IPAddressRange{}
			cerIpAddress.Min = split[0]
			cerIpAddress.Max = split[1]
			cerIpAddress.RangeStart, err = iputil.IpStrToHexString(cerIpAddress.Min, ipType)
			if err != nil {
				belogs.Error("ParseCerIpAddressModelByOpensslResults():IpNetToHexString min err:", err)
				return cerIpAddressModel, false, err
			}
			cerIpAddress.RangeEnd, err = iputil.IpStrToHexString(cerIpAddress.Max, ipType)
			if err != nil {
				belogs.Error("ParseCerIpAddressModelByOpensslResults():IpNetToHexString max err:", err)
				return cerIpAddressModel, false, err
			}
			addressPrefixRanges := iputil.IpRangeToAddressPrefixRanges(cerIpAddress.Min, cerIpAddress.Max)
			cerIpAddress.AddressPrefixRange = jsonutil.MarshalJson(addressPrefixRanges)
		} else {
			if len(ipp) == 0 || ipp == "inherit" {
				belogs.Debug("ParseCerIpAddressModelByOpensslResults():ipp is null or is inherit:", ipp)
				break
			}

			cerIpAddress.AddressPrefix, err = iputil.TrimAddressPrefixZero(ipp, ipType)
			if err != nil {
				belogs.Error("ParseCerIpAddressModelByOpensslResults():TrimAddressPrefixZero err:", err)
				return cerIpAddressModel, false, err
			}
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, err = iputil.AddressPrefixToHexRange(ipp, ipType)
			if err != nil {
				belogs.Error("ParseCerIpAddressModelByOpensslResults():AddressPrefixToHexRange err:", err)
				return cerIpAddressModel, false, err
			}
			cerIpAddress.AddressPrefixRange = jsonutil.MarshalJson(ipp)
		}
		cerIpAddressModel.CerIpAddresses = append(cerIpAddressModel.CerIpAddresses, cerIpAddress)
	}
	return cerIpAddressModel, false, nil
}

func ParseAsnModelByOpensslResults(results []string) (asnModel model.AsnModel, noAsn bool, err error) {
	/*
		sbgp-autonomousSysNum: critical
		      Autonomous System Numbers:
		        1-4294967295
		sbgp-autonomousSysNum: critical
		      Autonomous System Numbers:
		        inherit
	*/
	// AS
	asnModel.Asns = make([]model.Asn, 0)
	start := -1
	for i, one := range results {
		if strings.Contains(one, "Autonomous System Numbers:") {
			belogs.Debug("ParseAsnModelByOpensslResults(): one and results[i-1]:", one, results[i-1])
			if strings.Contains(results[i-1], "critical") {
				asnModel.Critical = true
			} else {
				asnModel.Critical = false
			}

			start = i + 1
			break
		}
	}
	if start < 0 {
		belogs.Debug("ParseAsnModelByOpensslResults(): no asn found")
		return asnModel, true, nil
	}
	ass := results[start:]
	for _, asTmp := range ass {
		as := strings.TrimSpace(asTmp)
		if len(as) == 0 || as == "inherit" {
			belogs.Debug("ParseAsnModelByOpensslResults():as is null or is inherit:", as)
			break
		}
		//default -1
		asn := model.NewAsn()
		split := strings.Split(as, "-")
		if len(split) == 2 {
			m, _ := strconv.Atoi(split[0])
			asn.Min = int64(m)
			m, _ = strconv.Atoi(split[1])
			asn.Max = int64(m)
		} else {
			m, _ := strconv.Atoi(as)
			asn.Asn = int64(m)
		}
		asnModel.Asns = append(asnModel.Asns, asn)

	}
	belogs.Debug("ParseAsnModelByOpensslResults(): asnModel", asnModel)
	return asnModel, false, nil
}
func ParseAiaModelSiaModelByOpensslResults(results []string) (aiaModel model.AiaModel, siaModel model.SiaModel, err error) {
	// AIA SIA

	for _, one := range results {
		if strings.Contains(one, "CA Issuers") {
			// AIA: Authority Information Access:
			tmp := strings.Replace(one, "CA Issuers - URI:", "", -1)
			aiaModel.CaIssuers = strings.TrimSpace(tmp)
		} else if strings.Contains(one, "Authority Information Access:") {
			if strings.Contains(one, "critical") {
				aiaModel.Critical = true
			} else {
				aiaModel.Critical = false
			}
		} else if strings.Contains(one, "CA Repository") {
			// SIA CaRepository
			tmp := strings.Replace(one, "CA Repository - URI:", "", -1)
			siaModel.CaRepository = strings.TrimSpace(tmp)
		} else if strings.Contains(one, "1.3.6.1.5.5.7.48.10") {
			// SIA rpkiMainfest
			tmp := strings.Replace(one, "1.3.6.1.5.5.7.48.10 - URI:", "", -1)
			siaModel.RpkiManifest = strings.TrimSpace(tmp)
		} else if strings.Contains(one, "1.3.6.1.5.5.7.48.13") {
			// SIA rpkiNotify
			tmp := strings.Replace(one, "1.3.6.1.5.5.7.48.13 - URI:", "", -1)
			siaModel.RpkiNotify = strings.TrimSpace(tmp)
		} else if strings.Contains(one, "1.3.6.1.5.5.7.48.11") {
			// SIA signedObject
			tmp := strings.Replace(one, "1.3.6.1.5.5.7.48.11 - URI:", "", -1)
			siaModel.SignedObject = strings.TrimSpace(tmp)
		} else if strings.Contains(one, "Subject Information Access:") {
			if strings.Contains(one, "critical") {
				siaModel.Critical = true
			} else {
				siaModel.Critical = false
			}
		}

	}
	return aiaModel, siaModel, nil
}

func ParseAkiSkiByOpensslResults(results []string) (aki, ski string, err error) {
	// AIA SIA
	for i, one := range results {
		if strings.Contains(one, "X509v3 Subject Key Identifier:") {
			belogs.Debug("ParseAkiSkiByOpensslResults(): SKI: results[i+1]: ", results[i+1])

			ski = strings.Replace(results[i+1], ":", "", -1)
			belogs.Debug("ParseAkiSkiByOpensslResults(): ski:", ski)

			ski = strings.ToLower(ski)
			ski = strings.TrimSpace(ski)
		}
		if strings.Contains(one, "X509v3 Authority Key Identifier:") {
			belogs.Debug("ParseAkiSkiByOpensslResults():AKI: results[i+1]: ", results[i+1])

			aki = strings.Replace(results[i+1], "keyid", "", -1)
			aki = strings.Replace(aki, ":", "", -1)
			aki = strings.ToLower(aki)
			aki = strings.TrimSpace(aki)
		}
	}
	return aki, ski, nil
}

func ParseSignatureAndPublicKeyByOpensslResults(results []string) (signatureInnerAlgorithm, signatureOuterAlgorithm model.Sha256RsaModel,
	publicKeyAlgorithm model.RsaModel, err error) {
	signatureOuterAlgorithmSha256Start := 0
	modulusStart := 0
	modulusEnd := 0
	for i, one := range results {
		if strings.Contains(one, "Signature Algorithm:") {
			if strings.Contains(results[i+1], "Issuer:") {
				signatureInnerAlgorithm.Name = strings.TrimSpace(strings.Replace(one, "Signature Algorithm:", "", -1))
			} else {
				signatureOuterAlgorithm.Name = strings.TrimSpace(strings.Replace(one, "Signature Algorithm:", "", -1))
				signatureOuterAlgorithmSha256Start = i + 1
			}

		}
		if strings.Contains(one, "Public Key Algorithm:") {
			publicKeyAlgorithm.Name = strings.TrimSpace(strings.Replace(one, "Public Key Algorithm:", "", -1))
		}
		if strings.Contains(one, "Modulus:") {
			modulusStart = i + 1
		}
		if strings.Contains(one, "Exponent:") {
			tmp := strings.Replace(one, "Exponent:", "", -1)
			split := strings.Split(tmp, "(")
			ex, _ := strconv.Atoi(strings.TrimSpace(split[0]))
			publicKeyAlgorithm.Exponent = uint64(ex)
			modulusEnd = i - 1
		}
	}
	modulus := bytes.NewBufferString("")
	for si := signatureOuterAlgorithmSha256Start; si < len(results); si++ {
		modulus.WriteString(strings.TrimSpace(results[si]))
	}
	signatureOuterAlgorithm.Sha256 = modulus.String()
	modulus = bytes.NewBufferString("")
	for si := modulusStart; si <= modulusEnd; si++ {
		modulus.WriteString(strings.TrimSpace(results[si]))
	}
	publicKeyAlgorithm.Modulus = modulus.String()
	return signatureInnerAlgorithm, signatureOuterAlgorithm, publicKeyAlgorithm, nil
}

func ParseKeyUsageModelByOpensslResults(results []string) (critical bool, KeyUsageValue string, err error) {

	for i, one := range results {
		if strings.Contains(one, "X509v3 Key Usage:") {
			if strings.Contains(one, "critical") {
				critical = true
			} else {
				critical = false
			}
			KeyUsageValue = strings.TrimSpace(results[i+1])
			break
		}
	}
	return critical, KeyUsageValue, nil
}

func ParseCrldpModelByOpensslResults(results []string) (critical bool, err error) {
	for _, one := range results {
		if strings.Contains(one, "X509v3 CRL Distribution Points:") {
			if strings.Contains(one, "critical") {
				critical = true
			} else {
				critical = false
			}
			break
		}
	}
	return critical, nil
}

func ParseBasicConstraintsModelByOpensslResults(results []string) (critical bool, err error) {
	for _, one := range results {
		if strings.Contains(one, "X509v3 Basic Constraints:") {
			if strings.Contains(one, "critical") {
				critical = true
			} else {
				critical = false
			}
			break
		}
	}
	return critical, nil
}
func ParseCertPolicyModelByOpensslResults(results []string) (certPolicyModel model.CertPolicyModel, err error) {
	for i, one := range results {
		if strings.Contains(one, "X509v3 Certificate Policies:") {
			if strings.Contains(one, "critical") {
				certPolicyModel.Critical = true
			} else {
				certPolicyModel.Critical = false
			}

			if strings.Contains(results[i+1], "1.3.6.1.5.5.7.14.2") {
				tmp := strings.Replace(results[i+2], "CPS:", "", -1)
				certPolicyModel.Cps = strings.TrimSpace(tmp)
			}
		}
	}
	return certPolicyModel, nil
}
