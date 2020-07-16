package model

import (
	"errors"
	"sync"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
)

type FileTypeIds struct {
	FileTypeIds []string `json:"fileTypeIds"`
}

func NewFileTypeIds(fileTypeId string) *FileTypeIds {
	fileTypeIds := &FileTypeIds{}
	fileTypeIds.FileTypeIds = make([]string, 0)
	fileTypeIds.FileTypeIds = append(fileTypeIds.FileTypeIds, fileTypeId)
	return fileTypeIds
}
func (c *FileTypeIds) Add(fileTypeId string) {
	c.FileTypeIds = append(c.FileTypeIds, fileTypeId)
}

// for to setup the al chains
type Chains struct {
	lock sync.RWMutex

	// key: fileTypeId,  value: chainCer, chainMft, chainRoa, chainClr
	FileTypeIdToCer map[string]ChainCer
	FileTypeIdToCrl map[string]ChainCrl
	FileTypeIdToMft map[string]ChainMft
	FileTypeIdToRoa map[string]ChainRoa

	// key: Aki, value: fileTypeId, may be more than one FileTypeId
	AkiToFileTypeIds map[string]FileTypeIds
	// key: Ski, value: fileTypeId
	SkiToFileTypeId map[string]string

	//store
	CerIds []uint64
	CrlIds []uint64
	MftIds []uint64
	RoaIds []uint64
}

func NewChains(count uint64) *Chains {
	chains := &Chains{}
	chains.CerIds = make([]uint64, 0)
	chains.CrlIds = make([]uint64, 0)
	chains.MftIds = make([]uint64, 0)
	chains.RoaIds = make([]uint64, 0)

	chains.FileTypeIdToCer = make(map[string]ChainCer, count)
	chains.FileTypeIdToCrl = make(map[string]ChainCrl, count)
	chains.FileTypeIdToMft = make(map[string]ChainMft, count)
	chains.FileTypeIdToRoa = make(map[string]ChainRoa, count)

	chains.AkiToFileTypeIds = make(map[string]FileTypeIds, count)
	chains.SkiToFileTypeId = make(map[string]string, count)
	return chains
}
func (c *Chains) AddCer(chainCer *ChainCer) {
	c.lock.Lock()
	defer c.lock.Unlock()

	fileTypeId := "cer" + convert.ToString(chainCer.Id)
	belogs.Debug("AddCer(): fileTypeId:", fileTypeId)

	// fileTypeId To Cer
	c.FileTypeIdToCer[fileTypeId] = *chainCer
	belogs.Debug("AddCer():add FileTypeIdToCer, fileTypeId , chainCer.Id:", fileTypeId, chainCer.Id)

	// Aki to fileTypeId
	fileTypeIds, ok := c.AkiToFileTypeIds[chainCer.Aki]
	belogs.Debug("AddCer():found AkiToFileTypeIds, chainCer.Aki:", chainCer.Aki, "   fileTypeId, ok:", fileTypeId, ok)
	if ok {
		fileTypeIds.Add(fileTypeId)
	} else {
		fileTypeIds = *NewFileTypeIds(fileTypeId)
	}
	c.AkiToFileTypeIds[chainCer.Aki] = fileTypeIds
	belogs.Debug("AddCer():add AkiToFileTypeIds, chainCer.Aki:", chainCer.Aki, "   len(fileTypeIds):", len(fileTypeIds.FileTypeIds))

	// ski to fileTypeId
	c.SkiToFileTypeId[chainCer.Ski] = fileTypeId
	belogs.Debug("AddCer():add SkiToFileTypeId, chainCer.Ski", chainCer.Ski, "  fileTypeId:", fileTypeId)

}

//
func (c *Chains) UpdateFileTypeIdToCer(chainCer *ChainCer) {
	c.lock.Lock()
	defer c.lock.Unlock()
	fileTypeId := "cer" + convert.ToString(chainCer.Id)
	c.FileTypeIdToCer[fileTypeId] = *chainCer
	belogs.Debug("UpdateFileTypeIdToCer():update FileTypeIdToCer, fileTypeId:", fileTypeId,
		"   chainCer.Id:", chainCer.Id, "   chainCer.StateModel:", chainCer.StateModel)
}

func (c *Chains) GetCerByFileTypeId(fileTypeId string) (chainCer ChainCer, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	chainCer, ok := c.FileTypeIdToCer[fileTypeId]
	if ok {
		belogs.Debug("GetCerByFileTypeId(): fileTypeId:", fileTypeId, "   chainCer.Id:", chainCer.Id,
			"   chainCer.StateModel:", chainCer.StateModel)
		return chainCer, nil
	}
	return chainCer, errors.New("not found chainCer by " + fileTypeId)

}

func (c *Chains) GetCerById(cerId uint64) (chainCer ChainCer, err error) {
	fileTypeId := "cer" + convert.ToString(cerId)
	return c.GetCerByFileTypeId(fileTypeId)
}

func (c *Chains) AddCrl(chainCrl *ChainCrl) {
	c.lock.Lock()
	defer c.lock.Unlock()

	fileTypeId := "crl" + convert.ToString(chainCrl.Id)
	belogs.Debug("AddCrl(): fileTypeId:", fileTypeId)
	// fileTypeId To Cer
	c.FileTypeIdToCrl[fileTypeId] = *chainCrl
	belogs.Debug("AddCrl():add FileTypeIdToCrl fileTypeId, chainCrl.Id:", fileTypeId, chainCrl.Id)
	// Aki to fileTypeId
	fileTypeIds, ok := c.AkiToFileTypeIds[chainCrl.Aki]
	belogs.Debug("AddCrl():found AkiToFileTypeIds, chainCrl.Aki,fileTypeId, ok", chainCrl.Aki, fileTypeId, ok)
	if ok {
		fileTypeIds.Add(fileTypeId)
	} else {
		fileTypeIds = *NewFileTypeIds(fileTypeId)
	}
	c.AkiToFileTypeIds[chainCrl.Aki] = fileTypeIds
	belogs.Debug("AddCrl():add AkiToFileTypeIds, chainCrl.Aki, len(fileTypeIds):", chainCrl.Aki, len(fileTypeIds.FileTypeIds))

	// no ski in crl
}
func (c *Chains) UpdateFileTypeIdToCrl(chainCrl *ChainCrl) {
	c.lock.Lock()
	defer c.lock.Unlock()
	fileTypeId := "crl" + convert.ToString(chainCrl.Id)
	c.FileTypeIdToCrl[fileTypeId] = *chainCrl
	belogs.Debug("UpdateFileTypeIdToCrl():update FileTypeIdToCrl, fileTypeId:", fileTypeId,
		"   chainCrl.Id:", chainCrl.Id, "   chainCrl.StateModel:", chainCrl.StateModel)
}
func (c *Chains) GetCrlByFileTypeId(fileTypeId string) (chainCrl ChainCrl, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	chainCrl, ok := c.FileTypeIdToCrl[fileTypeId]
	if ok {
		belogs.Debug("GetCrlByFileTypeId(): fileTypeId:", fileTypeId, "    chainCrl.Id:", chainCrl.Id,
			"   chainCrl.StateModel:", chainCrl.StateModel)
		return chainCrl, nil
	}
	return chainCrl, errors.New("not found chainCrl by " + fileTypeId)

}

func (c *Chains) GetCrlById(crlId uint64) (chainCrl ChainCrl, err error) {
	fileTypeId := "crl" + convert.ToString(crlId)
	return c.GetCrlByFileTypeId(fileTypeId)
}

func (c *Chains) AddMft(chainMft *ChainMft) {
	c.lock.Lock()
	defer c.lock.Unlock()

	fileTypeId := "mft" + convert.ToString(chainMft.Id)
	belogs.Debug("AddMft(): fileTypeId:", fileTypeId)

	// fileTypeId To Cer
	c.FileTypeIdToMft[fileTypeId] = *chainMft
	belogs.Debug("AddMft():add FileTypeIdToMft fileTypeId, chainMft.Id:", fileTypeId, chainMft.Id)

	// Aki to fileTypeId
	fileTypeIds, ok := c.AkiToFileTypeIds[chainMft.Aki]
	belogs.Debug("AddMft():found AkiToFileTypeIds, chainMft.Aki,fileTypeId, ok", chainMft.Aki, fileTypeId, ok)
	if ok {
		fileTypeIds.Add(fileTypeId)
	} else {
		fileTypeIds = *NewFileTypeIds(fileTypeId)
	}
	c.AkiToFileTypeIds[chainMft.Aki] = fileTypeIds
	belogs.Debug("AddMft():add AkiToFileTypeIds, chainMft.Aki, len(fileTypeIds):", chainMft.Aki, len(fileTypeIds.FileTypeIds))

	// ski to fileTypeId
	c.SkiToFileTypeId[chainMft.Ski] = fileTypeId
	belogs.Debug("AddMft():add SkiToFileTypeId, chainMft.Ski:fileTypeIds", chainMft.Ski, fileTypeIds)
}
func (c *Chains) UpdateFileTypeIdToMft(chainMft *ChainMft) {
	c.lock.Lock()
	defer c.lock.Unlock()
	fileTypeId := "mft" + convert.ToString(chainMft.Id)
	c.FileTypeIdToMft[fileTypeId] = *chainMft
	belogs.Debug("UpdateFileTypeIdToMft():update FileTypeIdToMft, fileTypeId:", fileTypeId,
		"   chainMft.Id:", chainMft.Id, "   chainMft.StateModel:", chainMft.StateModel)
}
func (c *Chains) GetMftByFileTypeId(fileTypeId string) (chainMft ChainMft, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	chainMft, ok := c.FileTypeIdToMft[fileTypeId]
	if ok {
		belogs.Debug("GetMftByFileTypeId(): fileTypeId:", fileTypeId, "   chainMft.Id, ok:", chainMft.Id,
			"   chainMft.StateModel:", chainMft.StateModel)
		return chainMft, nil
	}
	return chainMft, errors.New("not found chainMft by " + fileTypeId)

}
func (c *Chains) GetMftById(mftId uint64) (chainMft ChainMft, err error) {
	fileTypeId := "mft" + convert.ToString(mftId)
	return c.GetMftByFileTypeId(fileTypeId)
}

func (c *Chains) AddRoa(chainRoa *ChainRoa) {
	c.lock.Lock()
	defer c.lock.Unlock()

	fileTypeId := "roa" + convert.ToString(chainRoa.Id)
	belogs.Debug("AddRoa(): fileTypeId:", fileTypeId)

	// fileTypeId To Cer
	c.FileTypeIdToRoa[fileTypeId] = *chainRoa
	belogs.Debug("AddRoa():add FileTypeIdToRoa fileTypeId, chainRoa.Id:", fileTypeId, chainRoa.Id)
	// Aki to fileTypeId
	fileTypeIds, ok := c.AkiToFileTypeIds[chainRoa.Aki]
	belogs.Debug("AddRoa():found AkiToFileTypeIds, chainRoa.Aki, fileTypeId, ok", chainRoa.Aki, fileTypeId, ok)
	if ok {
		fileTypeIds.Add(fileTypeId)
	} else {
		fileTypeIds = *NewFileTypeIds(fileTypeId)
	}
	c.AkiToFileTypeIds[chainRoa.Aki] = fileTypeIds
	belogs.Debug("AddRoa():add AkiToFileTypeIds, chainRoa.Aki, len(fileTypeIds):", chainRoa.Aki, len(fileTypeIds.FileTypeIds))
	// ski to fileTypeId
	c.SkiToFileTypeId[chainRoa.Ski] = fileTypeId
	belogs.Debug("AddRoa():add SkiToFileTypeId, chainRoa.Ski:fileTypeIds:", chainRoa.Ski, fileTypeIds)
}
func (c *Chains) UpdateFileTypeIdToRoa(chainRoa *ChainRoa) {
	c.lock.Lock()
	defer c.lock.Unlock()
	fileTypeId := "roa" + convert.ToString(chainRoa.Id)
	c.FileTypeIdToRoa[fileTypeId] = *chainRoa
	belogs.Debug("UpdateFileTypeIdToRoa():update FileTypeIdToRoa, fileTypeId:", fileTypeId,
		"   chainRoa.Id:", chainRoa.Id, "   chainRoa.StateModel:", chainRoa.StateModel)
}

func (c *Chains) GetRoaByFileTypeId(fileTypeId string) (chainRoa ChainRoa, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	chainRoa, ok := c.FileTypeIdToRoa[fileTypeId]
	if ok {
		belogs.Debug("GetRoaByFileTypeId(): fileTypeId:", fileTypeId, "  chainRoa.Id, ok:", chainRoa.Id, ok)
		return chainRoa, nil
	}
	return chainRoa, errors.New("not found chainRoa by " + fileTypeId)

}

func (c *Chains) GetRoaById(roaId uint64) (chainRoa ChainRoa, err error) {
	fileTypeId := "roa" + convert.ToString(roaId)
	return c.GetRoaByFileTypeId(fileTypeId)
}
