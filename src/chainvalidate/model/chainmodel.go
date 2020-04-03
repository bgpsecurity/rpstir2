package model

import (
	"time"

	. "github.com/cpusoft/goutil/httpserver"

	"model"
)

// for chain validate
type ChainCerSql struct {
	Id        uint64    `json:"id" xorm:"id int"`
	FilePath  string    `json:"-" xorm:"filePath varchar(512)"`
	FileName  string    `json:"-" xorm:"fileName varchar(128)"`
	Ski       string    `json:"-" xorm:"ski varchar(128)"`
	Aki       string    `json:"-" xorm:"aki varchar(128)"`
	State     string    `json:"-" xorm:"state json"`
	IsRoot    bool      `json:"-" xorm:"isRoot"`
	NotBefore time.Time `json:"-" xorm:"notBefore datetime"`
	NotAfter  time.Time `json:"-" xorm:"notAfter datetime"`
}

func (c *ChainCerSql) ToChainCer() (chainCer ChainCer) {
	chainCer.Id = c.Id
	chainCer.FilePath = c.FilePath
	chainCer.FileName = c.FileName
	chainCer.Ski = c.Ski
	chainCer.Aki = c.Aki
	chainCer.State = c.State
	chainCer.IsRoot = c.IsRoot
	chainCer.NotBefore = c.NotBefore
	chainCer.NotAfter = c.NotAfter
	return chainCer
}

type ChainCer struct {
	Id        uint64    `json:"id" xorm:"id int"`
	FilePath  string    `json:"-" xorm:"filePath varchar(512)"`
	FileName  string    `json:"-" xorm:"fileName varchar(128)"`
	Ski       string    `json:"-" xorm:"ski varchar(128)"`
	Aki       string    `json:"-" xorm:"aki varchar(128)"`
	State     string    `json:"-" xorm:"state json"`
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
	Id            uint64 `json:"id" xorm:"id int"`
	AddressFamily uint64 `json:"-"  xorm:"addressFamily int"`
	//address prefix: 147.28.83.0/24 '
	AddressPrefix string `json:"-"  xorm:"addressPrefix varchar(512)"`
	MaxLength     uint64 `json:"-"  xorm:"maxLength int"`

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
	chainCrl.State = c.State
	return chainCrl
}

type ChainCrl struct {
	Id                uint64             `json:"id" xorm:"id int"`
	FilePath          string             `json:"-" xorm:"filePath varchar(512)"`
	FileName          string             `json:"-" xorm:"fileName varchar(128)"`
	Aki               string             `json:"-" xorm:"aki varchar(128)"`
	CrlNumber         uint64             `json:"-" xorm:"crlNumber int unsigned"`
	State             string             `json:"-" xorm:"state json"`
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

	Path string `json:"-"`
}
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
