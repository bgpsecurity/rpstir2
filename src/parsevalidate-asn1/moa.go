package parsevalidateasn1

import (
	"encoding/asn1"
	"os"
	"strings"
	"time"

	model "github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/byteutil"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

func ParseMoaModelByAsn1(fileModel *model.FileModel, moaModel *model.MoaModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("ParseMoaModelByAsn1(): will ReadFile", "fileModel", jsonutil.MarshalJson(fileModel))
	fileByte, err := os.ReadFile(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseMoaModelByAsn1(): ReadFile fail", "fileModel", jsonutil.MarshalJson(fileModel),
			"err", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	moaFileAsn1 := MoaFileAsn1{}
	_, err = asn1.Unmarshal(fileByte, &moaFileAsn1)
	if err != nil {
		// not add to stateModel
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Error("ParseMoaModelByAsn1(): Unmarshal moaFileAsn1 use asn1 fail, because indefinite length found,",
				"len(fileByte)", len(fileByte),
				"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		} else {
			belogs.Error("ParseMoaModelByAsn1(): Unmarshal moaFileAsn1 fail", "len(fileByte)", len(fileByte),
				"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
		}
		return err
	}

	belogs.Debug("ParseMoaModelByAsn1(): will range SignedDataAsn1s", "len(SignedDataAsn1s)", len(moaFileAsn1.SignedDataAsn1s))
	for _, seq := range moaFileAsn1.SignedDataAsn1s {
		belogs.Debug("ParseMoaModelByAsn1(): range SignedDataAsn1s", "seq.Tag", seq.Tag, "seq.Class", seq.Class,
			"seq.IsCompound", seq.IsCompound)

		if seq.Class == 0 && seq.Tag == 2 && !seq.IsCompound {
			// version CMSVersion INTEGER 3: ignore
		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) < 100 {
			// digestAlgorithms DigestAlgorithmIdentifiers SET (1 elem) : ignore
		} else if seq.Class == 0 && seq.Tag == 16 && seq.IsCompound {
			//  encapContentInfo EncapsulatedContentInfo
			eContentType, moaIpMapping, err := parseToMoaIpMapping(seq.FullBytes)
			if err != nil {
				belogs.Error("ParseMoaModelByAsn1(): parseToMoaIpMapping  fail",
					"len(seq.FullBytes)", len(seq.FullBytes),
					"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse moa by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				//continue
				return err
			}

			moaModel.EContentType = eContentType
			belogs.Debug("ParseMoaModelByAsn1(): parseToMoaIpMapping get eContentType",
				"moaModel.EContentType", moaModel.EContentType)

			if len(moaIpMapping.MoaMappingOctetStringAsn1) <= 1 {
				belogs.Error("ParseMoaModelByAsn1(): len(moaIpMapping.MoaMappingOctetStringAsn1)<1, fail",
					"len(seq.FullBytes)", len(seq.FullBytes),
					"len(moaIpMapping.MoaMappingOctetStringAsn1)", len(moaIpMapping.MoaMappingOctetStringAsn1),
					"fileModel", jsonutil.MarshalJson(fileModel))
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse moa by asn1",
					Detail: "parse moaIpMapping.MoaMappingOctetStringAsn1 fail"}
				stateModel.AddError(&stateMsg)
				//continue
				return err
			}
			moaMapping0 := moaIpMapping.MoaMappingOctetStringAsn1[0]
			belogs.Debug("ParseMoaModelByAsn1(): get moaIpMapping0",
				"moaMapping0.FullBytes", convert.PrintBytesOneLine(moaMapping0.FullBytes),
				"moaMapping0", jsonutil.MarshalJson(moaMapping0))

			var ipv6BitString asn1.BitString
			_, err = asn1.Unmarshal(moaMapping0.FullBytes, &ipv6BitString)
			if err != nil {
				// try again
				_, err = asn1.Unmarshal(moaMapping0.Bytes, &ipv6BitString)
				if err != nil {
					belogs.Error("ParseMoaModelByAsn1(): Unmarshal moaMapping0 fail",
						"moaMapping0.Bytes", convert.PrintBytesOneLine(moaMapping0.Bytes), "err", err)
					stateMsg := model.StateMsg{Stage: "parsevalidate",
						Fail:   "Fail to parse moa by asn1",
						Detail: err.Error()}
					stateModel.AddError(&stateMsg)
					return err
				}
			}
			ipv6Prefix, err := ParseBitStringToAddressPrefix(ipv6BitString, iputil.Ipv6Type)
			if err != nil {
				belogs.Error("ParseMoaModelByAsn1(): ParseBitStringToAddressPrefix ipv6BitString fail",
					"ipv6BitString", ipv6BitString, "err", err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse moa by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				return err
			}
			belogs.Debug("ParseMoaModelByAsn1(): get ipv6Prefix", "ipv6Prefix", ipv6Prefix)
			moaModel.Ipv6MappingPrefix = ipv6Prefix

			moaMapping1 := moaIpMapping.MoaMappingOctetStringAsn1[1]
			belogs.Debug("ParseMoaModelByAsn1(): get moaIpMapping1",
				"moaMapping1.FullBytes", convert.PrintBytesOneLine(moaMapping1.FullBytes),
				"moaMapping1", jsonutil.MarshalJson(moaMapping1))

			var ipv4BitStrings []asn1.BitString
			_, err = asn1.Unmarshal(moaMapping1.FullBytes, &ipv4BitStrings)
			if err != nil {
				// try again
				_, err = asn1.Unmarshal(moaMapping1.Bytes, &ipv4BitStrings)
				if err != nil {
					belogs.Error("ParseMoaModelByAsn1(): Unmarshal ipv4BitStrings fail",
						"moaMapping1.FullBytes", convert.PrintBytesOneLine(moaMapping1.FullBytes), "err", err)
					stateMsg := model.StateMsg{Stage: "parsevalidate",
						Fail:   "Fail to parse moa by asn1",
						Detail: err.Error()}
					stateModel.AddError(&stateMsg)
					return err
				}
			}
			ipv4Prefixs := make([]string, 0)
			var ipv4Prefix string
			for _, ipv4BitString := range ipv4BitStrings {
				ipv4Prefix, err = ParseBitStringToAddressPrefix(ipv4BitString, iputil.Ipv4Type)
				if err != nil {
					belogs.Error("ParseMoaModelByAsn1(): ParseBitStringToAddressPrefix ipv4BitString fail",
						"ipv4BitString", ipv4BitString, "err", err)
					stateMsg := model.StateMsg{Stage: "parsevalidate",
						Fail:   "Fail to parse moa by asn1",
						Detail: err.Error()}
					stateModel.AddError(&stateMsg)
					return err
				}
				belogs.Debug("ParseMoaModelByAsn1(): get ipv4Prefix", "ipv4Prefix", ipv4Prefix)
				ipv4Prefixs = append(ipv4Prefixs, ipv4Prefix)
			}
			belogs.Debug("ParseMoaModelByAsn1(): get ipv4Prefixs", "ipv4Prefixs", ipv4Prefixs)
			moaModel.Ipv4Prefixes = ipv4Prefixs

		} else if seq.Class == 2 && seq.Tag == 0 && seq.IsCompound {
			// EeModel
			var cerModel model.CerModel
			err = parseCerModelUseFileByteByAsn1(fileModel, seq.Bytes, true, &cerModel, stateModel)
			if err != nil {
				belogs.Error("ParseMoaModelByAsn1(): parseCerModelUseFileByteByAsn1 fail",
					"len(seq.Bytes)", len(seq.Bytes),
					"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			moaModel.SiaModel = cerModel.SiaModel
			moaModel.AiaModel = cerModel.AiaModel
			moaModel.Aki = cerModel.Aki
			moaModel.Ski = cerModel.Ski
			belogs.Debug("ParseMoaModelByAsn1(): parseCerModelUseFileByteByAsn1 get cerModel", "cerModel", cerModel.String())

			eeStart, eeEnd, err := byteutil.IndexStartAndEnd(fileByte, seq.Bytes)
			if err != nil {
				belogs.Error("ParseMoaModelByAsn1(): IndexStartAndEnd fail", "len(fileByte)", len(fileByte),
					"len(seq.Bytes)", len(seq.Bytes),
					"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseMoaModelByAsn1(): eeStart", eeStart, "eeEnd", eeEnd)
			moaModel.EeCertModel = *model.ConvertCertModelToEeCertModel(&cerModel, uint64(eeStart), uint64(eeEnd))
			belogs.Debug("ParseMoaModelByAsn1(): ConvertCertModelToEeCertModel get moaModel.EeModel",
				"moaModel.EeModel", jsonutil.MarshalJson(moaModel.EeCertModel))

		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) > 100 {
			moaModel.SignerInfoModel, err = ParseToSignerInfoModel(seq.Bytes)
			if err != nil {
				belogs.Error("ParseMoaModelByAsn1(): ParseToSignerInfoModel", "len(seq.Bytes)", len(seq.Bytes),
					"fileModel", jsonutil.MarshalJson(fileModel), "err", err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse signer by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseMoaModelByAsn1():ParseToSignerInfoModel get moaModel.SignerInfoModel",
				"moaModel.SignerInfoModel", jsonutil.MarshalJson(moaModel.SignerInfoModel))

		}
	}

	belogs.Info("ParseMoaModelByAsn1(): ok moaModel", "moaModel", moaModel.String(), "time(s)", time.Since(start))
	return
}

type MoaFileAsn1 struct {
	SignedDataOidAsn1 asn1.ObjectIdentifier `json:"signedDataOidAsn1"`
	SignedDataAsn1s   []asn1.RawValue       `json:"signedDataAsn1s" asn1:"optional,explicit,default:0,tag:0"`
}

type MoaOctetStringAsn1 struct {
	MoaOidAsn1         asn1.ObjectIdentifier `json:"moaOidAsn1"`
	MoaOctetStringAsn1 OctetString           `json:"moaOctetStringAsn1" asn1:"tag:0,explicit,optional"`
}

// asID as in rfc6482
type MoaIpMapping struct {
	Version                   int             `json:"version"`
	MoaMappingOctetStringAsn1 []asn1.RawValue // []asn1.BitString `json:"ipv4Prefixes"`
}

// data: asn1.FullBytes
func parseToMoaIpMapping(data []byte) (eContentType string, moaIpMapping MoaIpMapping, err error) {
	belogs.Debug("parseToMoaIpMapping(): in data", "len(data)", len(data))
	var moaOctetStringAsn1 MoaOctetStringAsn1
	_, err = asn1.Unmarshal(data, &moaOctetStringAsn1)
	if err != nil {
		belogs.Error("parseToMoaIpMapping(): Unmarshal to moaOctetStringAsn1 fail", "err", err)
		return
	}
	eContentType = moaOctetStringAsn1.MoaOidAsn1.String()
	belogs.Debug("parseToMoaIpMapping(): get eContentType",
		"moaOctetStringAsn1", jsonutil.MarshalJson(moaOctetStringAsn1),
		"moaOctetStringAsn1.MoaOctetStringAsn1",
		convert.PrintBytesOneLine(moaOctetStringAsn1.MoaOctetStringAsn1))

	_, err = asn1.Unmarshal([]byte(moaOctetStringAsn1.MoaOctetStringAsn1), &moaIpMapping)
	if err != nil {
		belogs.Error("parseToMoaIpMapping(): Unmarshal to moaIpMapping fail", "err", err)
		return
	}
	belogs.Debug("parseToMoaIpMapping(): get moaIpMapping.Version", "moaIpMapping.Version", moaIpMapping.Version,
		"moaIpMapping.MoaMappingOctetStringAsn1", moaIpMapping.MoaMappingOctetStringAsn1)

	return
}
