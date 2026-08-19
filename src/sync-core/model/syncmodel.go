package model

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
)

type SyncState struct {
	SyncStyle string `json:"syncStyle"`

	StartTime time.Time `json:"startTime,omitzero"`
	EndTime   time.Time `json:"endTime,omitzero"`

	SyncUrls   []string          `json:"syncUrls"`
	SyncResult *model.SyncResult `json:"syncResult"`
}

type LabDistributedSelectDetail struct {
	ID         int       `json:"id"`
	LogID      string    `json:"log_id"`
	Index      int       `json:"index"`
	StartTime  time.Time `json:"start_time"`
	SelectTime time.Time `json:"select_time"`
	Detail     string    `json:"detail"`
}
