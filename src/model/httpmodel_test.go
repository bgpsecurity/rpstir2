package model

import (
	"fmt"
	"testing"

	jsonutil "github.com/cpusoft/goutil/jsonutil"
)

func TestParseCert(t *testing.T) {
	body := `{"result":"ok","msg":"","certType":"cer","certModel":{"sn":"ab115a94034edb45","notBefore":"2017-09-14T19:04:19+08:00","notAfter":"2027-09-12T19:04:19+08:00","subject":"AfriNIC-Root-Certificate","issuer":"AfriNIC-Root-Certificate","ski":"eb680f38f5d6c71bb4b106b8bd06585012da31b6","aki":"","filePath":"/","fileName":"AfriNIC.cer","version":3,"basicConstraintsValid":true,"isRoot":true,"dnsNames":null,"emailAddresses":null,"ipAddresses":null,"subjectAll":"/CN=AfriNIC-Root-Certificate","issuerAll":"/CN=AfriNIC-Root-Certificate","state":"","crldpModels":[],"cerIpAddressModels":[{"addressPrefix":"0.0.0.0/0","min":"","max":"","rangeStart":"00.00.00.00","rangeEnd":"ff.ff.ff.ff"},{"addressPrefix":"::/0","min":"","max":"","rangeStart":"0000:0000:0000:0000:0000:0000:0000:0000","rangeEnd":"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"}],"asnModels":[{"asn":0,"min":0,"max":4294967295}],"aiaModel":{"caIssuers":""},"siaModel":{"rpkiManifest":"rsync://rpki.afrinic.net/repository/04E8B0D80F4D11E0B657D8931367AE7D/62gPOPXWxxu0sQa4vQZYUBLaMbY.mft","rpkiNotify":"","caRepository":"rsync://rpki.afrinic.net/repository/04E8B0D80F4D11E0B657D8931367AE7D/","signedObject":""}}}`
	fmt.Println(body)
	parseCertResponse := ParseCertResponse{}
	jsonutil.UnmarshalJson(string(body), &parseCertResponse)
	cerModel := CerModel{}
	jsonutil.UnmarshalJson(jsonutil.MarshalJson(parseCertResponse.CertModel), &cerModel)
	fmt.Println(cerModel)
}
