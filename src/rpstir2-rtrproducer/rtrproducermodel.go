package rtrproducer

type SerialNumberModel struct {
	SerialNumber        uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	GlobalSerialNumber  uint64 `json:"globalSerialNumber" xorm:"globalSerialNumber bigint"`
	SubpartSerialNumber uint64 `json:"subpartSerialNumber" xorm:"subpartSerialNumber bigint"`
}
type RushNodeIsTopResult struct {
	Id    uint64 `json:"id"`
	IsTop string `json:"isTop"`
}
