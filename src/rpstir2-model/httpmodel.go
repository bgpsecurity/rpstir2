package model

// certType: cer/crl/mft/roa
type ParseCertResponse struct {
	CertType   string      `json:"certType"`
	CertModel  interface{} `json:"certModel"`
	StateModel StateModel  `json:"stateModel"`
}

type TalModelsResponse struct {
	TalModels []TalModel `json:"talModels"`
}
