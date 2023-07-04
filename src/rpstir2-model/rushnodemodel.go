package model

import (
	"time"

	"github.com/guregu/null"
)

type RushNodeModel struct {
	Id       uint64 `json:"id" xorm:"id int"`
	NodeName string `json:"nodeName" xorm:"nodeName varchar(256)"`

	// if it is root, will be null
	ParentNodeId   null.Int `json:"parentNodeId" xorm:"parentNodeId int"`
	ParentNodeName string   `json:"parentNodeName" xorm:"parentNodeName varchar(256)"`
	// interface url: https://1.1.1.1:8080
	Url string `json:"url" xorm:"url varchar(256)"`
	// 'true/null: vc to identify itself. rp do not need this
	IsSelfUrl string `json:"isSelfUrl" xorm:"isSelfUrl varchar(8)"`

	// {"state":"valid"}, valid/invalid
	State string `json:"state" xorm:"state json"`
	Note  string `json:"note" xorm:"note varchar(256)"`
	//update time
	UpdateTime time.Time `json:"updateTime" xorm:"updateTime datetime"`
}
