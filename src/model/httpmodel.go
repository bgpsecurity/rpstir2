package model

import (
	. "github.com/cpusoft/goutil/httpserver"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
)

// certType: cer/crl/mft/roa
type ParseCertResponse struct {
	HttpResponse
	CertType   string      `json:"certType"`
	CertModel  interface{} `json:"certModel"`
	StateModel StateModel  `json:"stateModel"`
}
type SkiAki struct {
	Ski string `json:"ski"`
	Aki string `json:"aki"`
}

func GetModelFromJsonAll(jsonAll string) (certType string, certModel interface{}, err error) {
	parseCertResponse := ParseCertResponse{}
	jsonutil.UnmarshalJson(jsonAll, &parseCertResponse)
	certType = parseCertResponse.CertType
	certModelJson := jsonutil.MarshalJson(parseCertResponse.CertModel)

	switch certType {
	case "cer":
		cerModel := CerModel{}
		jsonutil.UnmarshalJson(certModelJson, &cerModel)
		certModel = cerModel
	case "roa":
		roaModel := RoaModel{}
		jsonutil.UnmarshalJson(certModelJson, &roaModel)
		certModel = roaModel
	case "crl":
		crlModel := CrlModel{}
		jsonutil.UnmarshalJson(certModelJson, &crlModel)
		certModel = crlModel
	case "mft":
		mftModel := MftModel{}
		jsonutil.UnmarshalJson(certModelJson, &mftModel)
		certModel = mftModel
	}
	return certType, certModel, nil

}

//
type StateResponse struct {
	HttpResponse
	State map[string]interface{} `json:"state"`
}

type ParseCertRepoResponse struct {
	HttpResponse
	CaRepository string `json:"caRepository"`
}
