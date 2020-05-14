package model

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
)

// for chaincert in db table
type ChainDbRoaModel struct {
	Id              uint64            `json:"id" xorm:"id int"`
	ParentChainCers []ChainDbCerModel `json:"parentChainCers,omitempty"`
}

func NewChainDbRoaModel(chainRoa *ChainRoa) *ChainDbRoaModel {
	chainDbRoaModel := &ChainDbRoaModel{}
	chainDbRoaModel.Id = chainRoa.Id

	chainDbRoaModel.ParentChainCers = make([]ChainDbCerModel, 0, len(chainRoa.ParentChainCerAlones))
	for i := range chainRoa.ParentChainCerAlones {
		// only save id
		chainDbCerModel := ChainDbCerModel{}
		chainDbCerModel.Id = chainRoa.ParentChainCerAlones[i].Id
		chainDbRoaModel.ParentChainCers = append(chainDbRoaModel.ParentChainCers, chainDbCerModel)
	}

	return chainDbRoaModel
}

type ChainDbCerModel struct {
	Id              uint64            `json:"id" xorm:"id int"`
	ParentChainCers []ChainDbCerModel `json:"parentChainCers,omitempty"`

	// child cer/crl/mft/roa ,just one level
	ChildChainCrls []ChainDbCrlModel `json:"childChainCrls,omitempty"`
	ChildChainMfts []ChainDbMftModel `json:"childChainMfts,omitempty"`
	ChildChainCers []ChainDbCerModel `json:"childChainCers,omitempty"`
	ChildChainRoas []ChainDbRoaModel `json:"childChainRoas,omitempty"`
}

func NewChainDbCerModel(chainCer *ChainCer) *ChainDbCerModel {
	chainDbCerModel := &ChainDbCerModel{}
	chainDbCerModel.Id = chainCer.Id

	chainDbCerModel.ChildChainCrls = make([]ChainDbCrlModel, 0, len(chainCer.ChildChainCrls))
	for i := range chainCer.ChildChainCrls {
		chainDbCrlModel := NewChainDbCrlModel(&chainCer.ChildChainCrls[i])
		chainDbCerModel.ChildChainCrls = append(chainDbCerModel.ChildChainCrls, *chainDbCrlModel)
	}
	belogs.Debug("NewChainDbCerModel():ChildChainCrls chainDbCerModel.Id:", chainDbCerModel.Id,
		"     len(chainCer.ChildChainCrls)", len(chainCer.ChildChainCrls),
		"     len(chainDbCerModel.ChildChainCrls):", len(chainDbCerModel.ChildChainCrls))

	chainDbCerModel.ChildChainMfts = make([]ChainDbMftModel, 0, len(chainCer.ChildChainMfts))
	for i := range chainCer.ChildChainMfts {
		chainDbMftModel := NewChainDbMftModel(&chainCer.ChildChainMfts[i])
		chainDbCerModel.ChildChainMfts = append(chainDbCerModel.ChildChainMfts, *chainDbMftModel)
	}
	belogs.Debug("NewChainDbCerModel():ChildChainMfts chainDbCerModel.Id:", chainDbCerModel.Id,
		"     len(chainCer.ChildChainMfts)", len(chainCer.ChildChainMfts),
		"     len(chainDbCerModel.ChildChainMfts):", len(chainDbCerModel.ChildChainMfts))

	chainDbCerModel.ChildChainCers = make([]ChainDbCerModel, 0, len(chainCer.ChildChainCerAlones))
	for i := range chainCer.ChildChainCerAlones {
		chainDbCerModelTmp := ChainDbCerModel{Id: chainCer.ChildChainCerAlones[i].Id}
		chainDbCerModel.ChildChainCers = append(chainDbCerModel.ChildChainCers, chainDbCerModelTmp)
	}
	belogs.Debug("NewChainDbCerModel():chainDbCerModel.Id:", chainDbCerModel.Id,
		"     len(chainCer.ChildChainCers)", len(chainCer.ChildChainCerAlones),
		"     len(chainDbCerModel.ChildChainCers):", len(chainDbCerModel.ChildChainCers))

	chainDbCerModel.ChildChainRoas = make([]ChainDbRoaModel, 0, len(chainCer.ChildChainRoas))
	for i := range chainCer.ChildChainRoas {
		chainDbRoaModel := ChainDbRoaModel{Id: chainCer.ChildChainRoas[i].Id}
		chainDbCerModel.ChildChainRoas = append(chainDbCerModel.ChildChainRoas, chainDbRoaModel)
	}
	belogs.Debug("NewChainDbCerModel():ChildChainRoas chainDbCerModel.Id:", chainDbCerModel.Id,
		"     len(chainCer.ChildChainRoas)", len(chainCer.ChildChainRoas),
		"     len(chainDbCerModel.ChildChainRoas):", len(chainDbCerModel.ChildChainRoas))

	chainDbCerModel.ParentChainCers = make([]ChainDbCerModel, 0)
	for i := range chainCer.ParentChainCerAlones {
		chainDbCerModelTmp := ChainDbCerModel{Id: chainCer.ParentChainCerAlones[i].Id}
		belogs.Debug("NewChainDbCerModel():i chainDbCerModel:", i, jsonutil.MarshalJson(chainDbCerModelTmp))
		chainDbCerModel.ParentChainCers = append(chainDbCerModel.ParentChainCers, chainDbCerModelTmp)
		belogs.Debug("NewChainDbCerModel():i, chainDbCerModel.ParentChainCers:", i, jsonutil.MarshalJson(chainDbCerModel.ParentChainCers))
	}
	belogs.Debug("NewChainDbCerModel():chainDbCerModel.Id:", chainDbCerModel.Id,
		"     len(chainCer.ParentChainCerAlones)", len(chainCer.ParentChainCerAlones),
		"     len(chainDbCerModel.ParentChainCers):", len(chainDbCerModel.ParentChainCers))
	return chainDbCerModel
}

type ChainDbCrlModel struct {
	Id              uint64            `json:"id" xorm:"id int"`
	ParentChainCers []ChainDbCerModel `json:"parentChainCers,omitempty"`
}

func NewChainDbCrlModel(chainCrl *ChainCrl) *ChainDbCrlModel {
	chainDbCrlModel := &ChainDbCrlModel{}
	chainDbCrlModel.Id = chainCrl.Id

	chainDbCrlModel.ParentChainCers = make([]ChainDbCerModel, 0, len(chainCrl.ParentChainCerAlones))
	for i := range chainCrl.ParentChainCerAlones {
		// only save id
		chainDbCerModel := ChainDbCerModel{}
		chainDbCerModel.Id = chainCrl.ParentChainCerAlones[i].Id
		chainDbCrlModel.ParentChainCers = append(chainDbCrlModel.ParentChainCers, chainDbCerModel)
	}
	return chainDbCrlModel
}

type ChainDbMftModel struct {
	Id              uint64            `json:"id" xorm:"id int"`
	ParentChainCers []ChainDbCerModel `json:"parentChainCers,omitempty"`
}

func NewChainDbMftModel(chainMft *ChainMft) *ChainDbMftModel {
	chainDbMftModel := &ChainDbMftModel{}
	chainDbMftModel.Id = chainMft.Id

	chainDbMftModel.ParentChainCers = make([]ChainDbCerModel, 0, len(chainMft.ParentChainCerAlones))
	for i := range chainMft.ParentChainCerAlones {
		// only save id
		chainDbCerModel := ChainDbCerModel{}
		chainDbCerModel.Id = chainMft.ParentChainCerAlones[i].Id
		chainDbMftModel.ParentChainCers = append(chainDbMftModel.ParentChainCers, chainDbCerModel)
	}
	return chainDbMftModel
}
