package parsevalidateasn1

import (
	"encoding/asn1"
	"encoding/hex"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/byteutil"
	"github.com/cpusoft/goutil/jsonutil"
)

type AsaFileAsn1 struct {
	SignedDataOidAsn1 asn1.ObjectIdentifier `json:"signedDataOidAsn1"`
	SignedDataAsn1s   []asn1.RawValue       `json:"signedDataAsn1s" asn1:"optional,explicit,default:0,tag:0"`
}

type AsaOctetStringAsn1 struct {
	AsaOidAsn1         asn1.ObjectIdentifier
	AsaOctetStringAsn1 OctetString `asn1:"tag:0,explicit,optional"`
}

// 1.2.840.113549.1.9.16.1.49
type AsaBlockAsn1 struct {
	//VersionAsn1      Version `json:"versionAsn1" asn1:"class:2,tag:0,optional"` //default 0
	CustomerAsnAsn1    int                 `json:"customerAsnAsn1"`
	ProviderBlockAsn1s []ProviderBlockAsn1 `json:"providerBlockAsn1s" asn1:"optional"`
}

type ProviderBlockAsn1 struct {
	ProviderAsnAsn1 int `json:"providerAsnAsn1"`
}

type AfiAsn1 struct {
	AfiAsn1 int
}
type AsaBlockOldAsn1 struct {
	AfiAsn1          AfiAsn1 `asn1:"class:2,tag:0"` //asn1.RawValue
	CustomerAsnAsn1  int     //asn1.RawValue   //`asn1:"explicit,tag:5"`
	ProviderAsnAsn1s []int   //`asn1:"explicit,tag:5"`
}

func convertAsaBlockOldAsn1ToAsaBlockAsn1(old AsaBlockOldAsn1) AsaBlockAsn1 {
	as := AsaBlockAsn1{}
	as.CustomerAsnAsn1 = old.CustomerAsnAsn1
	as.ProviderBlockAsn1s = make([]ProviderBlockAsn1, 0)
	for i := range old.ProviderAsnAsn1s {
		providerBlockAsn1 := ProviderBlockAsn1{
			ProviderAsnAsn1: old.ProviderAsnAsn1s[i],
		}
		as.ProviderBlockAsn1s = append(as.ProviderBlockAsn1s, providerBlockAsn1)
	}
	return as
}

func convertAsaBlockAsn1ToCustomerAsns(asaBlockAsn1 AsaBlockAsn1) (customerAsns []model.CustomerAsn, err error) {
	belogs.Debug("convertAsaBlockAsn1ToCustomerAsns(): asaBlockAsn1:", jsonutil.MarshalJson(asaBlockAsn1))

	customerAsns = make([]model.CustomerAsn, 0)
	customerAsn := model.CustomerAsn{}
	customerAsn.Version = uint64(0)
	customerAsn.CustomerAsn = uint64(asaBlockAsn1.CustomerAsnAsn1)
	providerAsns := make([]uint64, 0)
	for i := range asaBlockAsn1.ProviderBlockAsn1s {
		providerAsns = append(providerAsns, uint64(asaBlockAsn1.ProviderBlockAsn1s[i].ProviderAsnAsn1))
	}
	// 对 providerAsns 从小到大排序
	sort.Slice(providerAsns, func(i, j int) bool {
		return providerAsns[i] < providerAsns[j]
	})

	customerAsn.ProviderAsns = providerAsns
	customerAsns = append(customerAsns, customerAsn)
	// 对 customerAsns 按 CustomerAsn 从小到大排序
	sort.Slice(customerAsns, func(i, j int) bool {
		return customerAsns[i].CustomerAsn < customerAsns[j].CustomerAsn
	})

	belogs.Info("convertAsaBlockAsn1ToCustomerAsns(): customerAsns:", jsonutil.MarshalJson(customerAsns))
	return customerAsns, nil
}

// data: asn1.FullBytes
func parseToAsaBlockAsn1(data []byte) (eContentType string, asaBlockAsn1 AsaBlockAsn1, err error) {
	belogs.Debug("parseToAsaBlockAsn1(): len(data):", len(data))
	var asaOctetStringAsn1 AsaOctetStringAsn1
	_, err = asn1.Unmarshal(data, &asaOctetStringAsn1)
	if err != nil {
		belogs.Error("parseToAsaBlockAsn1(): Unmarshal to asaOctetStringAsn1 fail:", err)
		return
	}
	eContentType = asaOctetStringAsn1.AsaOidAsn1.String()
	belogs.Debug("parseToAsaBlockAsn1(): asaOctetStringAsn1:", jsonutil.MarshalJson(asaOctetStringAsn1),
		"    data:", hex.EncodeToString([]byte(asaOctetStringAsn1.AsaOctetStringAsn1)))

	asaBlockAsn1 = AsaBlockAsn1{}
	_, err = asn1.Unmarshal([]byte(asaOctetStringAsn1.AsaOctetStringAsn1), &asaBlockAsn1)
	if err != nil {
		belogs.Debug("parseToAsaBlockAsn1(): Unmarshal to asaBlockAsn1, try AsaBlockOldAsn1:", err)

		asaBlockOldAsn1 := AsaBlockOldAsn1{}
		_, err = asn1.Unmarshal([]byte(asaOctetStringAsn1.AsaOctetStringAsn1), &asaBlockOldAsn1)
		if err != nil {
			belogs.Error("parseToAsaBlockAsn1(): Unmarshal to asaBlockOldAsn1 fail:", hex.EncodeToString([]byte(asaOctetStringAsn1.AsaOctetStringAsn1)),
				err)
			return
		}
		belogs.Debug("parseToAsaBlockAsn1(): asaBlockOldAsn1:", jsonutil.MarshalJson(asaBlockOldAsn1))
		asaBlockAsn1 = convertAsaBlockOldAsn1ToAsaBlockAsn1(asaBlockOldAsn1)
	}
	belogs.Debug("parseToAsaBlockAsn1(): eContentType:", eContentType,
		"  asaBlockAsn1:", jsonutil.MarshalJson(asaBlockAsn1))
	return
}

func ParseAsaModelByAsn1(fileModel *model.FileModel, asaModel *model.AsaModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("ParseAsaModelByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel))
	fileByte, err := os.ReadFile(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseAsaModelByAsn1(): ReadFile fail, fileModel.TempFilePathName:", fileModel.TempFilePathName, err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("ParseAsaModelByAsn1(): ReadFile fileModel:", jsonutil.MarshalJson(fileModel), "  len(fileByte):", len(fileByte))
	asaFileAsn1 := AsaFileAsn1{}
	_, err = asn1.Unmarshal(fileByte, &asaFileAsn1)
	if err != nil {
		// not add to stateModel
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("ParseAsaModelByAsn1(): Unmarshal asaFileAsn1 use asn1 fail, because indefinite length found,",
				"   len(fileByte):", len(fileByte),
				"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("ParseAsaModelByAsn1(): Unmarshal asaFileAsn1 fail, len(fileByte):", len(fileByte),
				"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		return err
	}
	/*
		ParseAsaModelByAsn1(): seq.Tag: 2   seq.Class: 0   seq.IsCompound: false
		ParseAsaModelByAsn1(): seq.Tag: 17   seq.Class: 0   seq.IsCompound: true
		ParseAsaModelByAsn1(): seq.Tag: 16   seq.Class: 0   seq.IsCompound: true
		ParseAsaModelByAsn1(): seq.Tag: 0   seq.Class: 2   seq.IsCompound: true
		ParseAsaModelByAsn1(): seq.Tag: 17   seq.Class: 0   seq.IsCompound: true
	*/

	belogs.Debug("ParseAsaModelByAsn1(): len(SignedDataAsn1s):", len(asaFileAsn1.SignedDataAsn1s))
	for _, seq := range asaFileAsn1.SignedDataAsn1s {
		belogs.Debug("ParseAsaModelByAsn1(): seq.Tag:", seq.Tag, "  seq.Class:", seq.Class, "  seq.IsCompound:", seq.IsCompound)

		if seq.Class == 0 && seq.Tag == 2 && !seq.IsCompound {
			// version CMSVersion INTEGER 3: ignore
		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) < 100 {
			// digestAlgorithms DigestAlgorithmIdentifiers SET (1 elem) : ignore
		} else if seq.Class == 0 && seq.Tag == 16 && seq.IsCompound {
			// //  encapContentInfo EncapsulatedContentInfo

			eContentType, asaBlockAsn1, err := parseToAsaBlockAsn1(seq.FullBytes)
			if err != nil {
				belogs.Error("ParseAsaModelByAsn1(): parseToAsaBlockAsn1 fail, len(seq.FullBytes):", len(seq.FullBytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file to get aspa by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			asaModel.EContentType = eContentType
			belogs.Debug("ParseAsaModelByAsn1(): parseToAsaBlockAsn1 asaModel.EContentType:", asaModel.EContentType,
				"  asaBlockAsn1:", jsonutil.MarshalJson(asaBlockAsn1))

			asaModel.CustomerAsns, err = convertAsaBlockAsn1ToCustomerAsns(asaBlockAsn1)
			if err != nil {
				belogs.Error("ParseAsaModelByAsn1():convertAsaBlockAsn1ToCustomerAsns err, asaBlockAsn1:", jsonutil.MarshalJson(asaBlockAsn1),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file to get aspa by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseAsaModelByAsn1(): convertAsaBlockAsn1ToCustomerAsns asaModel.EContentType:", asaModel.EContentType,
				"  asaModel.CustomerAsns:", jsonutil.MarshalJson(asaModel.CustomerAsns))

		} else if seq.Class == 2 && seq.Tag == 0 && seq.IsCompound {
			// EeModel will call
			var cerModel model.CerModel
			err = parseCerModelUseFileByteByAsn1(fileModel, seq.Bytes, true, &cerModel, stateModel)
			if err != nil {
				belogs.Error("ParseAsaModelByAsn1(): parseCerModelUseFileByteByAsn1 fail, len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file to get ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			asaModel.SiaModel = cerModel.SiaModel
			asaModel.AiaModel = cerModel.AiaModel
			asaModel.Aki = cerModel.Aki
			asaModel.Ski = cerModel.Ski
			belogs.Debug("ParseAsaModelByAsn1(): parseCerModelUseFileByteByAsn1 cerModel:", cerModel.String())

			eeCertStart, eeCertEnd, err := byteutil.IndexStartAndEnd(fileByte, seq.Bytes)
			if err != nil {
				belogs.Error("ParseAsaModelByAsn1(): IndexStartAndEnd fail, len(fileByte):", len(fileByte),
					"   len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file to get ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseAsaModelByAsn1(): eeCertStart:", eeCertStart, "  eeCertEnd:", eeCertEnd)
			asaModel.EeCertModel = *model.ConvertCertModelToEeCertModel(&cerModel, uint64(eeCertStart), uint64(eeCertEnd))
			belogs.Debug("ParseAsaModelByAsn1(): ConvertCertModelToEeCertModel asaModel.EeCertModel:", jsonutil.MarshalJson(asaModel.EeCertModel))

		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) > 100 {
			// signerInfos SignerInfos will call
			asaModel.SignerInfoModel, err = ParseToSignerInfoModel(seq.Bytes)
			if err != nil {
				belogs.Error("ParseAsaModelByAsn1(): ParseToSignerInfoModel, len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file to get signer by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseAsaModelByAsn1():ParseToSignerInfoModel  asaModel.SignerInfoModel:", jsonutil.MarshalJson(asaModel.SignerInfoModel))

		}
	}
	belogs.Debug("ParseAsaModelByAsn1(): ok fileModel:", jsonutil.MarshalJson(fileModel), "  time(s):", time.Since(start))
	return
}
