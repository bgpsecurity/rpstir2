package model

import (
	. "github.com/cpusoft/goutil/httpserver"
)

// certType: cer/crl/mft/roa
type ParseCertResponse struct {
	HttpResponse
	CertType   string      `json:"certType"`
	CertModel  interface{} `json:"certModel"`
	StateModel StateModel  `json:"stateModel"`
}

//
type StateResponse struct {
	HttpResponse
	State map[string]interface{} `json:"state"`
}

type ParseCerSimpleResponse struct {
	HttpResponse
	ParseCerSimple ParseCerSimple `json:"parseCerSimple"`
}
type TalResponse struct {
	HttpResponse
	TalModels []TalModel `json:"talModels"`
}
type RsyncResultResponse struct {
	HttpResponse
	RsyncResult SyncResult `json:"rsyncResult"`
}
