package common

type SerialNumberModel struct {
	SerialNumber        uint64 `json:"serialNumber" xorm:"serialNumber bigint"`
	GlobalSerialNumber  uint64 `json:"globalSerialNumber" xorm:"globalSerialNumber bigint"`
	SubpartSerialNumber uint64 `json:"subpartSerialNumber" xorm:"subpartSerialNumber bigint"`
	// when roa or asa, will insert to lab_rpki_rtr_serial_number using goroutine
	HaveSaveToDb uint32 `json:"-"`
}
