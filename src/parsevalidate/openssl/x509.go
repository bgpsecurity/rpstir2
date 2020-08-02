package openssl

import (
	"crypto/x509"
	"fmt"
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	"parsevalidate/util"
)

func ParseCerModelByX509(fileByte []byte, cerModel *model.CerModel) (err error) {
	// cert , use x509 to get value
	cer, err := x509.ParseCertificate(fileByte)
	if err != nil {
		belogs.Error("ParseCerModelByX509():ParseCertificate fail: len(fileByte):", len(fileByte), err)
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
	cerModel.ExtKeyUsages = util.ExtKeyUsagesToInts(cer.ExtKeyUsage)
	cerModel.IsCa = cer.IsCA

	//SHA256-RSA
	//cerModel.SignatureAlgorithm = cer.SignatureAlgorithm.String()
	//RSA
	//cerModel.PublicKeyAlgorithm = cer.PublicKeyAlgorithm.String()

	//	fmt.Printf("serialNumber=%s\r\n", cert.Subject.SerialNumber)
	//	fmt.Printf("serialNumber=%s\r\n", cert.Issuer.SerialNumber)
	//	fmt.Printf("SN=%v\r\n", cert.SerialNumber.Uint64())

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
	//AIA
	//cerModel.Aia = cert.IssuingCertificateURL

	cerModel.ExtensionModels = make([]model.ExtensionModel, 0)
	for _, ext := range cer.Extensions {
		extensionModel := model.ExtensionModel{
			Oid:      ext.Id.String(),
			Critical: ext.Critical,
		}
		if name, ok := model.CerExtensionOids[ext.Id.String()]; ok {
			extensionModel.Name = name
		}
		cerModel.ExtensionModels = append(cerModel.ExtensionModels, extensionModel)
	}

	belogs.Debug("ParseCerModelByX509():cerModel:", jsonutil.MarshalJson(cerModel))
	return nil
}

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
	eeCertModel.SubjectAll, _ = util.GetDNFromName(cer.Subject, ",")
	eeCertModel.IssuerAll, _ = util.GetDNFromName(cer.Issuer, ",")
	eeCertModel.KeyUsageModel.KeyUsage = int(cer.KeyUsage)
	eeCertModel.ExtKeyUsages = util.ExtKeyUsagesToInts(cer.ExtKeyUsage)

	//CRLDPS
	eeCertModel.CrldpModel.Crldps = make([]string, 0)
	for _, crldp := range cer.CRLDistributionPoints {
		eeCertModel.CrldpModel.Crldps = append(eeCertModel.CrldpModel.Crldps, crldp)
	}
	return nil
}

func ParseCrlModelByX509(fileByte []byte, crlModel *model.CrlModel) (err error) {

	crl, err := x509.ParseCRL(fileByte)
	if err != nil {
		belogs.Error("ParseCrlModelByX509():ParseCRL err:", err)
		return err
	}
	belogs.Debug("ParseCrlModelByX509():crl:", jsonutil.MarshalJson(crl))

	tbsCertList := crl.TBSCertList
	crlModel.Version = tbsCertList.Version
	crlModel.IssuerAll, _ = util.GetDNFromRDNSeq(tbsCertList.Issuer, ",")
	crlModel.ThisUpdate = tbsCertList.ThisUpdate.Local()
	crlModel.NextUpdate = tbsCertList.NextUpdate.Local()
	crlModel.HasExpired = strconv.FormatBool(crl.HasExpired(time.Now()))
	//exts := tbsCertList.Extensions
	crlModel.RevokedCertModels = make([]model.RevokedCertModel, 0)
	revokedCerts := tbsCertList.RevokedCertificates
	for _, revokedCert := range revokedCerts {
		revokedCertModel := model.RevokedCertModel{}
		revokedCertModel.Sn = fmt.Sprintf("%x", revokedCert.SerialNumber)
		revokedCertModel.RevocationTime = revokedCert.RevocationTime.Local()
		crlModel.RevokedCertModels = append(crlModel.RevokedCertModels, revokedCertModel)
	}
	belogs.Debug("ParseCrlModelByX509():crlModel:", jsonutil.MarshalJson(crlModel))

	return nil
}
