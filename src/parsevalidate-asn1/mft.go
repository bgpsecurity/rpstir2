package parsevalidateasn1

import (
	"encoding/asn1"
	"encoding/hex"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/byteutil"
	"github.com/cpusoft/goutil/jsonutil"
)

type MftFileAsn1 struct {
	SignedDataOidAsn1 asn1.ObjectIdentifier `json:"signedDataOidAsn1"`
	SignedDataAsn1s   []asn1.RawValue       `json:"signedDataAsn1s" asn1:"optional,explicit,default:0,tag:0"`
}

type MftOctetStringAsn1 struct {
	MftOidAsn1         asn1.ObjectIdentifier `json:"mftOidAsn1"`
	MftOctetStringAsn1 OctetString           `json:"mftOctetStringAsn1" asn1:"tag:0,explicit,optional"`
}

// asID as in rfc6482
type MftBlockAsn1 struct {
	MftNumberAsn1         *big.Int              `json:"mftNumberAsn1"`
	ThisUpdateAsn1        time.Time             `asn1:"generalized" json:"thisUpdateAsn1"`
	NextUpdateAsn1        time.Time             `asn1:"generalized" json:"nextUpdateAsn1"`
	FileHashAlgorithmAsn1 asn1.ObjectIdentifier `json:"fileHashAlgorithmAsn1"`
	FileAndHashAsn1s      []FileAndHashAsn1     `json:"fileAndHashAsn1s"`
}
type FileAndHashAsn1 struct {
	FileAsn1 string         `asn1:"ia5" json:"fileAsn1"`
	HashAsn1 asn1.BitString `json:"hashAsn1"`
}

// data: asn1.FullBytes
func parseToMftBlockAsn1(data []byte) (eContentType string, mftBlockAsn1 MftBlockAsn1, err error) {
	belogs.Debug("parseToMftBlockAsn1(): len(data):", len(data))
	var mftOctetStringAsn1 MftOctetStringAsn1
	_, err = asn1.Unmarshal(data, &mftOctetStringAsn1)
	if err != nil {
		belogs.Error("parseToMftBlockAsn1(): Unmarshal to mftOctetStringAsn1 fail:", err)
		return
	}
	eContentType = mftOctetStringAsn1.MftOidAsn1.String()
	belogs.Debug("parseToMftBlockAsn1(): mftOctetStringAsn1:", jsonutil.MarshalJson(mftOctetStringAsn1))

	_, err = asn1.Unmarshal([]byte(mftOctetStringAsn1.MftOctetStringAsn1), &mftBlockAsn1)
	if err != nil {
		belogs.Error("parseToMftBlockAsn1(): Unmarshal to mftBlockAsn1 fail:", err)
		return
	}
	belogs.Debug("parseToMftBlockAsn1():mftBlockAsn1:", jsonutil.MarshalJson(mftBlockAsn1))
	return
}

func ParseMftModelByAsn1(fileModel *model.FileModel, mftModel *model.MftModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("ParseMftModelByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel))
	fileByte, err := os.ReadFile(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseMftModelByAsn1(): ReadFile fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file by asn1",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	mftFileAsn1 := MftFileAsn1{}
	_, err = asn1.Unmarshal(fileByte, &mftFileAsn1)
	if err != nil {
		// not add to stateModel
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("ParseMftModelByAsn1(): Unmarshal mftFileAsn1 use asn1 fail, because indefinite length found,",
				"   len(fileByte):", len(fileByte),
				"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("ParseMftModelByAsn1(): Unmarshal mftFileAsn1 fail, len(fileByte):", len(fileByte),
				"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		return err
	}
	/*
	   ParseMftModelByAsn1(): seq.Tag: 2   seq.Class: 0   seq.IsCompound: false
	   ParseMftModelByAsn1(): seq.Tag: 17   seq.Class: 0   seq.IsCompound: true
	   ParseMftModelByAsn1(): seq.Tag: 16   seq.Class: 0   seq.IsCompound: true
	   ParseMftModelByAsn1(): seq.Tag: 0   seq.Class: 2   seq.IsCompound: true
	   ParseMftModelByAsn1(): seq.Tag: 17   seq.Class: 0   seq.IsCompound: true
	*/

	belogs.Debug("ParseMftModelByAsn1(): len(SignedDataAsn1s):", len(mftFileAsn1.SignedDataAsn1s))
	for _, seq := range mftFileAsn1.SignedDataAsn1s {
		belogs.Debug("ParseMftModelByAsn1(): seq.Tag:", seq.Tag, "  seq.Class:", seq.Class, "  seq.IsCompound:", seq.IsCompound)

		if seq.Class == 0 && seq.Tag == 2 && !seq.IsCompound {
			// version CMSVersion INTEGER 3: ignore
		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) < 100 {
			// digestAlgorithms DigestAlgorithmIdentifiers SET (1 elem) : ignore
		} else if seq.Class == 0 && seq.Tag == 16 && seq.IsCompound {
			// //  encapContentInfo EncapsulatedContentInfo
			eContentType, mftBlockAsn1, err := parseToMftBlockAsn1(seq.FullBytes)
			if err != nil {
				belogs.Error("ParseMftModelByAsn1(): parseToMftBlockAsn1 fail, len(seq.FullBytes):", len(seq.FullBytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse mft by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			mftModel.EContentType = eContentType
			belogs.Debug("ParseMftModelByAsn1(): parseToMftBlockAsn1 mftModel.EContentType:", mftModel.EContentType,
				"  mftBlockAsn1:", jsonutil.MarshalJson(mftBlockAsn1))

			mftModel.MftNumber = mftBlockAsn1.MftNumberAsn1.String()
			mftModel.ThisUpdate = mftBlockAsn1.ThisUpdateAsn1
			mftModel.NextUpdate = mftBlockAsn1.NextUpdateAsn1
			mftModel.FileHashAlg = mftBlockAsn1.FileHashAlgorithmAsn1.String()
			mftModel.FileHashModels = make([]model.FileHashModel, 0)
			for i := range mftBlockAsn1.FileAndHashAsn1s {
				fileHashModel := model.FileHashModel{
					File: mftBlockAsn1.FileAndHashAsn1s[i].FileAsn1,
					Hash: hex.EncodeToString(mftBlockAsn1.FileAndHashAsn1s[i].HashAsn1.Bytes),
				}
				mftModel.FileHashModels = append(mftModel.FileHashModels, fileHashModel)
			}
			belogs.Debug("ParseMftModelByAsn1(): parseToMftBlockAsn1 mftModel.FileHashModels:", jsonutil.MarshalJson(mftModel.FileHashModels))

		} else if seq.Class == 2 && seq.Tag == 0 && seq.IsCompound {
			// EeModel will call
			// EeModel
			var cerModel model.CerModel
			err = parseCerModelUseFileByteByAsn1(fileModel, seq.Bytes, true, &cerModel, stateModel)
			if err != nil {
				belogs.Error("ParseMftModelByAsn1(): parseCerModelUseFileByteByAsn1 fail, len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			mftModel.SiaModel = cerModel.SiaModel
			mftModel.AiaModel = cerModel.AiaModel
			mftModel.Aki = cerModel.Aki
			mftModel.Ski = cerModel.Ski
			belogs.Debug("ParseMftModelByAsn1(): parseCerModelUseFileByteByAsn1 cerModel:", cerModel.String())

			eeCertStart, eeCertEnd, err := byteutil.IndexStartAndEnd(fileByte, seq.Bytes)
			if err != nil {
				belogs.Error("ParseMftModelByAsn1(): IndexStartAndEnd fail, len(fileByte):", len(fileByte),
					"   len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseMftModelByAsn1(): eeCertStart:", eeCertStart, "  eeCertEnd:", eeCertEnd)
			mftModel.EeCertModel = *model.ConvertCertModelToEeCertModel(&cerModel, uint64(eeCertStart), uint64(eeCertEnd))
			belogs.Debug("ParseMftModelByAsn1(): ConvertCertModelToEeCertModel mftModel.EeCertModel:", jsonutil.MarshalJson(mftModel.EeCertModel))

		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) > 100 {
			// signerInfos SignerInfos will call
			mftModel.SignerInfoModel, err = ParseToSignerInfoModel(seq.Bytes)
			if err != nil {
				belogs.Error("ParseMftModelByAsn1(): ParseToSignerInfoModel, len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse signer by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseMftModelByAsn1():ParseToSignerInfoModel  mftModel.SignerInfoModel:", jsonutil.MarshalJson(mftModel.SignerInfoModel))
		}
	}
	belogs.Info("ParseMftModelByAsn1(): ok, mftModel:", mftModel.String(), " time(s):", time.Since(start))
	return
}
