package model

import (
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	. "github.com/cpusoft/goutil/httpserver"
	"github.com/cpusoft/goutil/jsonutil"

	"model"
)

// for chain validate
type ChainCertSql struct {
	Id             uint64    `json:"id" xorm:"id int"`
	JsonAll        string    `json:"-" xorm:"jsonAll json"`
	State          string    `json:"-" xorm:"state json"`
	CrlFileName    string    `json:"-" xorm:"crlFileName varchar(128)"`
	RevocationTime time.Time `json:"-" xorm:"revocationTime datetime"`

	// for crl
	CerFiles string `json:"-" xorm:"cerFiles varchar(2048)"`
	RoaFiles string `json:"-" xorm:"roaFiles varchar(2048)"`
	MftFiles string `json:"-" xorm:"mftFiles varchar(2048)"`
}

func (c *ChainCertSql) ToChainCer() (chainCer ChainCer) {
	chainCer.Id = c.Id

	cerModel := model.CerModel{}
	err := jsonutil.UnmarshalJson(c.JsonAll, &cerModel)
	belogs.Debug("ToChainCer(): cerModel, err:", jsonutil.MarshalJson(cerModel), err)

	chainCer.FilePath = cerModel.FilePath
	chainCer.FileName = cerModel.FileName
	chainCer.Ski = cerModel.Ski
	chainCer.Aki = cerModel.Aki
	chainCer.IsRoot = cerModel.IsRoot
	chainCer.NotBefore = cerModel.NotBefore
	chainCer.NotAfter = cerModel.NotAfter

	cerIpAddress := jsonutil.MarshalJson(cerModel.CerIpAddressModel.CerIpAddresses)
	belogs.Debug("ToChainCer(): cerIpAddress:", cerIpAddress)
	jsonutil.UnmarshalJson(cerIpAddress, &chainCer.ChainIpAddresses)
	belogs.Debug("ToChainCer(): chainCer.ChainIpAddresses:", chainCer.ChainIpAddresses)

	asns := jsonutil.MarshalJson(cerModel.AsnModel.Asns)
	belogs.Debug("ToChainCer(): asns:", asns)
	jsonutil.UnmarshalJson(asns, &chainCer.ChainAsns)
	belogs.Debug("ToChainCer(): chainCer.ChainAsns:", chainCer.ChainAsns)

	chainCer.StateModel = model.GetStateModelAndResetStage(c.State, "chainvalidate")
	chainCer.ChainSnInCrlRevoked = ChainSnInCrlRevoked{
		CrlFileName: c.CrlFileName, RevocationTime: c.RevocationTime}
	belogs.Debug("ToChainCer(): chainCer:", chainCer)
	return chainCer
}
func (c *ChainCertSql) ToChainCrl() (chainCrl ChainCrl) {
	chainCrl.Id = c.Id

	crlModel := model.CrlModel{}
	err := jsonutil.UnmarshalJson(c.JsonAll, &crlModel)
	belogs.Debug("ToChainCrl(): crlModel, err:", jsonutil.MarshalJson(crlModel), err)

	chainCrl.FilePath = crlModel.FilePath
	chainCrl.FileName = crlModel.FileName
	chainCrl.Aki = crlModel.Aki
	chainCrl.CrlNumber = crlModel.CrlNumber

	revokedCertModels := jsonutil.MarshalJson(crlModel.RevokedCertModels)
	belogs.Debug("ToChainCrl(): revokedCertModels:", revokedCertModels)
	jsonutil.UnmarshalJson(revokedCertModels, &chainCrl.ChainRevokedCerts)
	belogs.Debug("ToChainCrl(): chainCrl.ChainRevokedCerts:", chainCrl.ChainRevokedCerts)

	belogs.Debug("ToChainCrl(): c.CerFiles:", c.CerFiles, "   c.RoaFiles:", c.RoaFiles, "   c.MftFiles:", c.MftFiles)
	shouldRevokedCerts := make([]string, 0)
	if len(c.CerFiles) > 0 {
		cerFiles := strings.Split(c.CerFiles, ",")
		shouldRevokedCerts = append(shouldRevokedCerts, cerFiles...)
	}
	if len(c.RoaFiles) > 0 {
		roaFiles := strings.Split(c.RoaFiles, ",")
		shouldRevokedCerts = append(shouldRevokedCerts, roaFiles...)
	}
	if len(c.MftFiles) > 0 {
		mftFiles := strings.Split(c.MftFiles, ",")
		shouldRevokedCerts = append(shouldRevokedCerts, mftFiles...)
	}

	chainCrl.ShouldRevokedCerts = shouldRevokedCerts
	belogs.Debug("ToChainCrl(): chainCrl.ShouldRevokedCerts:", chainCrl.ShouldRevokedCerts, "    len(chainCrl.ShouldRevokedCerts):", len(chainCrl.ShouldRevokedCerts))

	chainCrl.StateModel = model.GetStateModelAndResetStage(c.State, "chainvalidate")
	belogs.Debug("ToChainCrl(): chainCrl:", chainCrl)
	return chainCrl
}

func (c *ChainCertSql) ToChainMft() (chainMft ChainMft) {
	chainMft.Id = c.Id

	mftModel := model.MftModel{}
	err := jsonutil.UnmarshalJson(c.JsonAll, &mftModel)
	belogs.Debug("ToChainMft(): mftModel, err:", jsonutil.MarshalJson(mftModel), err)

	chainMft.FilePath = mftModel.FilePath
	chainMft.FileName = mftModel.FileName
	chainMft.Aki = mftModel.Aki
	chainMft.Ski = mftModel.Ski
	chainMft.MftNumber = mftModel.MftNumber
	chainMft.EeCertStart = mftModel.EeCertModel.EeCertStart
	chainMft.EeCertEnd = mftModel.EeCertModel.EeCertEnd

	chainMft.StateModel = model.GetStateModelAndResetStage(c.State, "chainvalidate")
	belogs.Debug("ToChainMft(): chainMft:", chainMft)
	return chainMft
}

func (c *ChainCertSql) ToChainRoa() (chainRoa ChainRoa) {
	chainRoa.Id = c.Id

	roaModel := model.RoaModel{}
	err := jsonutil.UnmarshalJson(c.JsonAll, &roaModel)
	belogs.Debug("ToChainRoa(): roaModel, err:", jsonutil.MarshalJson(roaModel), err)

	chainRoa.FilePath = roaModel.FilePath
	chainRoa.FileName = roaModel.FileName
	chainRoa.Ski = roaModel.Ski
	chainRoa.Aki = roaModel.Aki
	chainRoa.EeCertStart = roaModel.EeCertModel.EeCertStart
	chainRoa.EeCertEnd = roaModel.EeCertModel.EeCertEnd

	chainIpAddresses := jsonutil.MarshalJson(roaModel.RoaIpAddressModels)
	belogs.Debug("ToChainRoa(): chainIpAddresses:", chainIpAddresses)
	jsonutil.UnmarshalJson(chainIpAddresses, &chainRoa.ChainIpAddresses)
	belogs.Debug("ToChainRoa(): chainRoa.ChainIpAddresses:", chainRoa.ChainIpAddresses)

	chainEeIpAddresses := jsonutil.MarshalJson(roaModel.EeCertModel.CerIpAddressModel.CerIpAddresses)
	belogs.Debug("ToChainRoa(): chainEeIpAddresses:", chainEeIpAddresses)
	jsonutil.UnmarshalJson(chainEeIpAddresses, &chainRoa.ChainEeIpAddresses)
	belogs.Debug("ToChainRoa(): chainRoa.ChainEeIpAddresses:", chainRoa.ChainEeIpAddresses)

	chainRoa.StateModel = model.GetStateModelAndResetStage(c.State, "chainvalidate")
	chainRoa.ChainSnInCrlRevoked = ChainSnInCrlRevoked{
		CrlFileName: c.CrlFileName, RevocationTime: c.RevocationTime}
	belogs.Debug("ToChainRoa(): chainRoa:", chainRoa)
	return chainRoa
}

type ChainCer struct {
	Id        uint64    `json:"id" xorm:"id int"`
	FilePath  string    `json:"-" xorm:"filePath varchar(512)"`
	FileName  string    `json:"-" xorm:"fileName varchar(128)"`
	Ski       string    `json:"-" xorm:"ski varchar(128)"`
	Aki       string    `json:"-" xorm:"aki varchar(128)"`
	IsRoot    bool      `json:"-" xorm:"isRoot"`
	NotBefore time.Time `json:"-" xorm:"notBefore datetime"`
	NotAfter  time.Time `json:"-" xorm:"notAfter datetime"`

	// all ip address
	ChainIpAddresses []ChainIpAddress `json:"-"`

	// all asn
	ChainAsns []ChainAsn `json:"-"`

	//all parent cer, trace back to root
	ParentChainCerAlones []ChainCerAlone `json:"parentChainCers"`

	//should be revoked
	ChainSnInCrlRevoked ChainSnInCrlRevoked `json:"-"`

	//,omitempty
	// child cer/crl/mft/roa ,just one level
	ChildChainCerAlones []ChainCerAlone `json:"childChainCerIds"`
	ChildChainCrls      []ChainCrl      `json:"childChainCrls"`
	ChildChainMfts      []ChainMft      `json:"childChainMfts"`
	ChildChainRoas      []ChainRoa      `json:"childChainRoas"`

	StateModel model.StateModel `json:"-"`
}

// just like ChainCer, but no parents ,no children
// avoid circulation
type ChainCerAlone struct {
	Id       uint64 `json:"id" xorm:"id int"`
	FilePath string `json:"-" xorm:"filePath varchar(512)"`
	FileName string `json:"-" xorm:"fileName varchar(128)"`
	IsRoot   bool   `json:"-" xorm:"isRoot"`
	// all ip address
	ChainIpAddresses []ChainIpAddress `json:"-"`
	// all asn
	ChainAsns []ChainAsn `json:"-"`
}

func NewChainCerAlone(chainCer *ChainCer) (chainCerAlone *ChainCerAlone) {
	chainCerAlone = &ChainCerAlone{
		Id:               chainCer.Id,
		FilePath:         chainCer.FilePath,
		FileName:         chainCer.FileName,
		IsRoot:           chainCer.IsRoot,
		ChainIpAddresses: chainCer.ChainIpAddresses,
		ChainAsns:        chainCer.ChainAsns,
	}
	return chainCerAlone

}

type ChainIpAddress struct {
	Id uint64 `json:"id" xorm:"id int"`
	//AddressFamily uint64 `json:"-"  xorm:"addressFamily int"`
	//address prefix: 147.28.83.0/24 '
	//AddressPrefix string `json:"-"  xorm:"addressPrefix varchar(512)"`
	//MaxLength     uint64 `json:"-"  xorm:"maxLength int"`

	//min address range from addressPrefix or min/max, in hex:  63.60.00.00'
	RangeStart string `json:"rangeStart" xorm:"rangeStart varchar(512)"`
	//max address range from addressPrefix or min/max, in hex:  63.69.7f.ff'
	RangeEnd string `json:"rangeEnd" xorm:"rangeEnd varchar(512)"`
}

type ChainAsn struct {
	Id  uint64 `json:"id" xorm:"id int"`
	Asn uint64 `json:"asn" xorm:"asn int"`
	Min uint64 `json:"min" xorm:"min int"`
	Max uint64 `json:"max" xorm:"max int"`
}

type ChainSnInCrlRevoked struct {
	CrlFileName    string    `json:"-" xorm:"fileName varchar(512)"`
	RevocationTime time.Time `json:"-" xorm:"revocationTime datetime"`
}

/*
type ChainCrlSql struct {
	Id        uint64 `json:"id" xorm:"id int"`
	FilePath  string `json:"-" xorm:"filePath varchar(512)"`
	FileName  string `json:"-" xorm:"fileName varchar(128)"`
	Aki       string `json:"-" xorm:"aki varchar(128)"`
	CrlNumber uint64 `json:"-" xorm:"crlNumber int unsigned"`
	State     string `json:"-" xorm:"state json"`
}

func (c *ChainCrlSql) ToChainCrl() (chainCrl ChainCrl) {
	chainCrl.Id = c.Id
	chainCrl.FilePath = c.FilePath
	chainCrl.FileName = c.FileName
	chainCrl.Aki = c.Aki
	chainCrl.CrlNumber = c.CrlNumber
	//chainCrl.State = c.State
	return chainCrl
}
*/
type ChainCrl struct {
	Id        uint64 `json:"id" xorm:"id int"`
	FilePath  string `json:"-" xorm:"filePath varchar(512)"`
	FileName  string `json:"-" xorm:"fileName varchar(128)"`
	Aki       string `json:"-" xorm:"aki varchar(128)"`
	CrlNumber uint64 `json:"-" xorm:"crlNumber int unsigned"`

	ChainRevokedCerts []ChainRevokedCert `json:"-"`

	// certs(cer, roa, mft) by sn, should not exists
	ShouldRevokedCerts []string `json:"-"`

	// parent cer
	ParentChainCerAlones []ChainCerAlone `json:"parentChainCers"`

	StateModel model.StateModel `json:"-"`
}
type ChainRevokedCert struct {
	Sn string `json:"-" xorm:"sn varchar(512)"`
}

/*
type ChainMftSql struct {
	Id          uint64 `json:"id" xorm:"id int"`
	FilePath    string `json:"-" xorm:"filePath varchar(512)"`
	FileName    string `json:"-" xorm:"fileName varchar(128)"`
	Ski         string `json:"-" xorm:"ski varchar(128)"`
	Aki         string `json:"-" xorm:"aki varchar(128)"`
	MftNumber   string `json:"-" xorm:"mftNumber varchar(1024)"`
	State       string `json:"-" xorm:"state json"`
	EeCertStart uint64 `json:"-" xorm:"eeCertStart int"`
	EeCertEnd   uint64 `json:"-" xorm:"eeCertEnd int"`
}

func (c *ChainMftSql) ToChainMft() (chainMft ChainMft) {
	chainMft.Id = c.Id
	chainMft.FilePath = c.FilePath
	chainMft.FileName = c.FileName
	chainMft.Ski = c.Ski
	chainMft.Aki = c.Aki
	chainMft.MftNumber = c.MftNumber
	chainMft.State = c.State
	chainMft.EeCertStart = c.EeCertStart
	chainMft.EeCertEnd = c.EeCertEnd
	return chainMft
}
*/
type ChainMft struct {
	Id             uint64          `json:"id" xorm:"id int"`
	FilePath       string          `json:"-" xorm:"filePath varchar(512)"`
	FileName       string          `json:"-" xorm:"fileName varchar(128)"`
	Ski            string          `json:"-" xorm:"ski varchar(128)"`
	Aki            string          `json:"-" xorm:"aki varchar(128)"`
	MftNumber      string          `json:"-" xorm:"mftNumber varchar(1024)"`
	State          string          `json:"-" xorm:"state json"`
	EeCertStart    uint64          `json:"-" xorm:"eeCertStart int"`
	EeCertEnd      uint64          `json:"-" xorm:"eeCertEnd int"`
	ChainFileHashs []ChainFileHash `json:"-"`

	// parent cer
	ParentChainCerAlones []ChainCerAlone `json:"parentChainCers"`

	//should be revoked
	ChainSnInCrlRevoked ChainSnInCrlRevoked `json:"-"`

	StateModel model.StateModel `json:"-"`
}
type ChainFileHash struct {
	File string `json:"-" xorm:"file varchar(1024)"`
	Hash string `json:"-" xorm:"hash varchar(1024)"`

	Path string `json:"-" xorm:"path varchar(2048)"`
}

/*
type ChainRoaSql struct {
	Id          uint64 `json:"id" xorm:"id int"`
	Asn         uint64 `json:"-" xorm:"asn int"`
	FilePath    string `json:"-" xorm:"filePath varchar(512)"`
	FileName    string `json:"-" xorm:"fileName varchar(128)"`
	Ski         string `json:"-" xorm:"ski varchar(128)"`
	Aki         string `json:"-" xorm:"aki varchar(128)"`
	State       string `json:"-" xorm:"state json"`
	EeCertStart uint64 `json:"-" xorm:"eeCertStart int"`
	EeCertEnd   uint64 `json:"-" xorm:"eeCertEnd int"`
}

func (c *ChainRoaSql) ToChainRoa() (chainRoa ChainRoa) {
	chainRoa.Id = c.Id
	chainRoa.Asn = c.Asn
	chainRoa.FilePath = c.FilePath
	chainRoa.FileName = c.FileName
	chainRoa.Ski = c.Ski
	chainRoa.Aki = c.Aki
	chainRoa.State = c.State
	chainRoa.EeCertStart = c.EeCertStart
	chainRoa.EeCertEnd = c.EeCertEnd
	return chainRoa
}
*/
type ChainRoa struct {
	Id          uint64 `json:"id" xorm:"id int"`
	Asn         uint64 `json:"-" xorm:"asn int"`
	FilePath    string `json:"-" xorm:"filePath varchar(512)"`
	FileName    string `json:"-" xorm:"fileName varchar(128)"`
	Ski         string `json:"-" xorm:"ski varchar(128)"`
	Aki         string `json:"-" xorm:"aki varchar(128)"`
	State       string `json:"-" xorm:"state json"`
	EeCertStart uint64 `json:"-" xorm:"eeCertStart int"`
	EeCertEnd   uint64 `json:"-" xorm:"eeCertEnd int"`

	// parent cer
	ParentChainCerAlones []ChainCerAlone `json:"parentChainCers"`

	//should be revoked
	ChainSnInCrlRevoked ChainSnInCrlRevoked `json:"-"`

	// all ip address
	ChainIpAddresses []ChainIpAddress `json:"-"`
	// all ee address
	ChainEeIpAddresses []ChainIpAddress `json:"-"`

	StateModel model.StateModel `json:"-"`
}

type ChainValidateFileRequest struct {
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
}

type ChainValidateFileResponse struct {
	HttpResponse
	//ChainCer/ChainCrl/ChainRoa/ChainMft/
	ChainCert  interface{}      `json:"chainCert"`
	StateModel model.StateModel `json:"stateModel"`
}
