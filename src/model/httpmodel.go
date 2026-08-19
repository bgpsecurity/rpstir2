package model

// certType: cer/crl/mft/roa
type ParseCertResponse struct {
	CertType    string      `json:"certType"`
	FileHash    string      `json:"fileHash"`
	CertModel   interface{} `json:"certModel"`
	OriginModel OriginModel `json:"originModel"`
	StateModel  StateModel  `json:"stateModel"`
}

type TalModelsResponse struct {
	TalModels []TalModel `json:"talModels"`
}
