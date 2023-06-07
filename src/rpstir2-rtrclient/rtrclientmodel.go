package rtrclient

type RtrClientStartModel struct {
	Server string `json:"server"`
	Port   string `json:"port"`
}

type RtrClientSerialQueryModel struct {
	SessionId    uint16 `json:"sessionId"`
	SerialNumber uint32 `json:"serialNumber"`
}
