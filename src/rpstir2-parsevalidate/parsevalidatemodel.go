package parsevalidate

import "time"

type CertIdStateModel struct {
	Id       uint64 `json:"id" xorm:"id int"`
	StateStr string `json:"stateStr" xorm:"stateStr varchar"`
	// nextUpdate, or notAfter
	EndTime time.Time `json:"endTime" xorm:"endTime datetime"`
}
