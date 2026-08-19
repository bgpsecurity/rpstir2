package parsevalidateasn1

import (
	"crypto/x509"
	"fmt"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/asn1util"
	"github.com/cpusoft/goutil/belogs"
)

func ParseEeCertModelByX509(fileByte []byte, eeCertModel *model.EeCertModel) (err error) {
	// cert
	belogs.Debug("ParseEeCertModelByX509():len(fileByte):", len(fileByte))
	//logs.LogDebugBytes("ParseEeCertModelByX509():(fileByte):", (fileByte))
	cer, err := x509.ParseCertificate(fileByte)
	if err != nil {
		belogs.Error("ParseEeCertModelByX509():ParseCertificate err:", err)
		return err
	}

	eeCertModel.Sn = fmt.Sprintf("%x", cer.SerialNumber)
	eeCertModel.Version = cer.Version
	eeCertModel.DigestAlgorithm = cer.SignatureAlgorithm.String()
	eeCertModel.NotBefore = cer.NotBefore.Local()

	eeCertModel.NotAfter = cer.NotAfter.Local()
	eeCertModel.SubjectAll, _ = asn1util.GetDNFromName(cer.Subject, ",")
	eeCertModel.IssuerAll, _ = asn1util.GetDNFromName(cer.Issuer, ",")
	eeCertModel.KeyUsageModel.KeyUsage = int(cer.KeyUsage)
	eeCertModel.ExtKeyUsages = asn1util.ExtKeyUsagesToInts(cer.ExtKeyUsage)

	//CRLDPS
	eeCertModel.CrldpModel.Crldps = make([]string, 0)
	for _, crldp := range cer.CRLDistributionPoints {
		eeCertModel.CrldpModel.Crldps = append(eeCertModel.CrldpModel.Crldps, crldp)
	}
	return nil
}
