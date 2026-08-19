package parsevalidateasn1

import (
	"encoding/asn1"
	"os"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/byteutil"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

type RoaFileAsn1 struct {
	SignedDataOidAsn1 asn1.ObjectIdentifier `json:"signedDataOidAsn1"`
	SignedDataAsn1s   []asn1.RawValue       `json:"signedDataAsn1s" asn1:"optional,explicit,default:0,tag:0"`
}

type RoaOctetStringAsn1 struct {
	RoaOidAsn1         asn1.ObjectIdentifier `json:"roaOidAsn1"`
	RoaOctetStringAsn1 OctetString           `json:"roaOctetStringAsn1" asn1:"tag:0,explicit,optional"`
}

// asID as in rfc6482
type RoaBlockAsn1 struct {
	RoaAsIdAsn1                RoaAsIdAsn1                 `json:"roaAsIdAsn1"`
	RoaAddressFamilyBlockAsn1s []RoaAddressFamilyBlockAsn1 `json:"roaAddressFamilyBlockAsn1"`
}
type RoaAsIdAsn1 int64
type RoaAddressFamilyBlockAsn1 struct {
	RoaAddressFamilyAsn1 []byte                `json:"roaAddressFamilyAsn1"`
	RoaAddressBlockAsn1s []RoaAddressBlockAsn1 `json:"roaAddressBlockAsn1s"`
}
type RoaAddressBlockAsn1 struct {
	RoaAddressPrefixAsn1 asn1.BitString `json:"roaAddressPrefixAsn1"`
	RoaMaxLengthAsn1     int            `json:"roaMaxLengthAsn1" asn1:"optional,default:-1"`
}

type DigestAlgorithm struct {
	Oid  asn1.ObjectIdentifier
	Null asn1.RawValue
}

// data: asn1.FullBytes
func parseToRoaBlockAsn1(data []byte) (eContentType string, roaBlockAsn1 RoaBlockAsn1, err error) {
	belogs.Debug("parseToRoaBlockAsn1(): len(data):", len(data))
	var roaOctetStringAsn1 RoaOctetStringAsn1
	_, err = asn1.Unmarshal(data, &roaOctetStringAsn1)
	if err != nil {
		belogs.Error("parseToRoaBlockAsn1(): Unmarshal to roaOctetStringAsn1 fail:", err)
		return
	}
	eContentType = roaOctetStringAsn1.RoaOidAsn1.String()
	belogs.Debug("parseToRoaBlockAsn1(): roaOctetStringAsn1:", jsonutil.MarshalJson(roaOctetStringAsn1))

	_, err = asn1.Unmarshal([]byte(roaOctetStringAsn1.RoaOctetStringAsn1), &roaBlockAsn1)
	if err != nil {
		belogs.Error("parseToRoaBlockAsn1(): Unmarshal to roaBlockAsn1 fail:", err)
		return
	}
	belogs.Debug("parseToRoaBlockAsn1():roaBlockAsn1:", jsonutil.MarshalJson(roaBlockAsn1))
	return
}

func ParseRoaModelByAsn1(fileModel *model.FileModel, roaModel *model.RoaModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("ParseRoaModelByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel))
	fileByte, err := os.ReadFile(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseRoaModelByAsn1(): ReadFile fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	roaFileAsn1 := RoaFileAsn1{}
	_, err = asn1.Unmarshal(fileByte, &roaFileAsn1)
	if err != nil {
		// not add to stateModel
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("ParseRoaModelByAsn1(): Unmarshal roaFileAsn1 use asn1 fail, because indefinite length found,",
				"   len(fileByte):", len(fileByte),
				"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("ParseRoaModelByAsn1(): Unmarshal roaFileAsn1 fail, len(fileByte):", len(fileByte),
				"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		return err
	}
	roaIpAddressModels := make([]model.RoaIpAddressModel, 0)
	belogs.Debug("ParseRoaModelByAsn1(): len(SignedDataAsn1s):", len(roaFileAsn1.SignedDataAsn1s))
	for _, seq := range roaFileAsn1.SignedDataAsn1s {
		belogs.Debug("ParseRoaModelByAsn1(): seq.Tag:", seq.Tag, "  seq.Class:", seq.Class, "  seq.IsCompound:", seq.IsCompound)

		if seq.Class == 0 && seq.Tag == 2 && !seq.IsCompound {
			// version CMSVersion INTEGER 3: ignore
		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) < 100 {
			// digestAlgorithms DigestAlgorithmIdentifiers SET (1 elem) : ignore
		} else if seq.Class == 0 && seq.Tag == 16 && seq.IsCompound {
			//  encapContentInfo EncapsulatedContentInfo
			eContentType, roaBlockAsn1, err := parseToRoaBlockAsn1(seq.FullBytes)
			if err != nil {
				belogs.Error("ParseRoaModelByAsn1(): parseToRoaBlockAsn1 fail, len(seq.FullBytes):", len(seq.FullBytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse roa by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}

			roaModel.Asn = int64(roaBlockAsn1.RoaAsIdAsn1)
			roaModel.EContentType = eContentType
			belogs.Debug("ParseRoaModelByAsn1(): parseToRoaBlockAsn1 roaModel.EContentType:", roaModel.EContentType, " roaModel.Asn:", roaModel.Asn)

			belogs.Debug("ParseRoaModelByAsn1(): len(roaBlockAsn1.RoaAddressFamilyBlockAsn1s):", len(roaBlockAsn1.RoaAddressFamilyBlockAsn1s))
			for i := range roaBlockAsn1.RoaAddressFamilyBlockAsn1s {
				roaAddressFamilyBlockAsn1 := roaBlockAsn1.RoaAddressFamilyBlockAsn1s[i]
				belogs.Debug("ParseRoaModelByAsn1(): range roaBlockAsn1.RoaAddressFamilyBlockAsn1s, i:", i, " len(roaAddressFamilyBlockAsn1.RoaAddressBlockAsn1s):", len(roaAddressFamilyBlockAsn1.RoaAddressBlockAsn1s))

				for j := range roaAddressFamilyBlockAsn1.RoaAddressBlockAsn1s {
					roaIpAddressModel := model.RoaIpAddressModel{}
					roaIpAddressModel.AddressFamily = convert.BytesToBigInt(roaAddressFamilyBlockAsn1.RoaAddressFamilyAsn1).Uint64()
					belogs.Debug("ParseRoaModelByAsn1(): roaIpAddressModel.AddressFamily:", roaIpAddressModel.AddressFamily)

					roaAddressBlockAsn1 := roaAddressFamilyBlockAsn1.RoaAddressBlockAsn1s[j]
					belogs.Debug("ParseRoaModelByAsn1(): roaAddressBlockAsn1:", jsonutil.MarshalJson(roaAddressBlockAsn1))

					maxLength := int(roaAddressBlockAsn1.RoaMaxLengthAsn1)
					if maxLength > 0 {
						roaIpAddressModel.MaxLength = uint64(maxLength)
					} else {
						roaIpAddressModel.MaxLength = uint64(roaAddressBlockAsn1.RoaAddressPrefixAsn1.BitLength)
					}
					belogs.Debug("ParseRoaModelByAsn1(): roaIpAddressModel.MaxLength:", roaIpAddressModel.MaxLength)

					roaIpAddressModel.AddressPrefix, err = ParseBitStringToAddressPrefix(roaAddressBlockAsn1.RoaAddressPrefixAsn1,
						int(roaIpAddressModel.AddressFamily))
					if err != nil {
						belogs.Error("ParseRoaModelByAsn1(): ParseToAddressPrefix fail, roaAddressBlockAsn1:", roaAddressBlockAsn1,
							"   fileModel:", jsonutil.MarshalJson(fileModel), err)
						stateMsg := model.StateMsg{Stage: "parsevalidate",
							Fail:   "Fail to parse roa's address by asn1",
							Detail: err.Error()}
						stateModel.AddError(&stateMsg)
						continue
					}
					belogs.Debug("ParseRoaModelByAsn1(): roaIpAddressModel.AddressPrefix:", roaIpAddressModel.AddressPrefix,
						"   roaIpAddressModel.MaxLength:", roaIpAddressModel.MaxLength)

					roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, err =
						iputil.AddressPrefixToHexRange(roaIpAddressModel.AddressPrefix, int(roaIpAddressModel.AddressFamily))
					if err != nil {
						belogs.Error("ParseRoaModelByAsn1(): AddressPrefixToHexRange fail, addressprefix:", roaIpAddressModel.AddressPrefix,
							"   addressFamily:", roaIpAddressModel.AddressFamily,
							"   fileModel:", jsonutil.MarshalJson(fileModel), err)
						stateMsg := model.StateMsg{Stage: "parsevalidate",
							Fail:   "Fail to parse roa's address by asn1",
							Detail: err.Error()}
						stateModel.AddError(&stateMsg)
						continue
					}
					belogs.Debug("ParseRoaModelByAsn1(): roaIpAddressModel.RangeStart:", roaIpAddressModel.RangeStart,
						"   roaIpAddressModel.RangeEnd:", roaIpAddressModel.RangeEnd)

					addressPrefixWithZero, err := iputil.FillAddressPrefixWithZero(roaIpAddressModel.AddressPrefix,
						int(roaIpAddressModel.AddressFamily))
					if err != nil {
						belogs.Error("ParseRoaModelByAsn1(): FillAddressPrefixWithZero fail,  addressprefix:", roaIpAddressModel.AddressPrefix,
							"   addressFamily:", roaIpAddressModel.AddressFamily,
							"   fileModel:", jsonutil.MarshalJson(fileModel), err)
						stateMsg := model.StateMsg{Stage: "parsevalidate",
							Fail:   "Fail to parse roa's address by asn1",
							Detail: err.Error()}
						stateModel.AddError(&stateMsg)
						continue
					}
					roaIpAddressModel.AddressPrefixRange = jsonutil.MarshalJson(addressPrefixWithZero)
					belogs.Debug("ParseRoaModelByAsn1(): addressPrefixWithZero:", addressPrefixWithZero,
						"   roaIpAddressModel.AddressPrefixRange:", roaIpAddressModel.AddressPrefixRange)
					roaIpAddressModels = append(roaIpAddressModels, roaIpAddressModel)
				}

			}
		} else if seq.Class == 2 && seq.Tag == 0 && seq.IsCompound {
			// EeModel
			var cerModel model.CerModel
			err = parseCerModelUseFileByteByAsn1(fileModel, seq.Bytes, true, &cerModel, stateModel)
			if err != nil {
				belogs.Error("ParseRoaModelByAsn1(): parseCerModelUseFileByteByAsn1 fail, len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			roaModel.SiaModel = cerModel.SiaModel
			roaModel.AiaModel = cerModel.AiaModel
			roaModel.Aki = cerModel.Aki
			roaModel.Ski = cerModel.Ski
			belogs.Debug("ParseRoaModelByAsn1(): parseCerModelUseFileByteByAsn1 cerModel:", cerModel.String())

			eeCertStart, eeCertEnd, err := byteutil.IndexStartAndEnd(fileByte, seq.Bytes)
			if err != nil {
				belogs.Error("ParseRoaModelByAsn1(): IndexStartAndEnd fail, len(fileByte):", len(fileByte),
					"   len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse ee by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseRoaModelByAsn1(): eeCertStart:", eeCertStart, "  eeCertEnd:", eeCertEnd)
			roaModel.EeCertModel = *model.ConvertCertModelToEeCertModel(&cerModel, uint64(eeCertStart), uint64(eeCertEnd))
			belogs.Debug("ParseRoaModelByAsn1(): ConvertCertModelToEeCertModel roaModel.EeCertModel:", jsonutil.MarshalJson(roaModel.EeCertModel))

		} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) > 100 {
			roaModel.SignerInfoModel, err = ParseToSignerInfoModel(seq.Bytes)
			if err != nil {
				belogs.Error("ParseRoaModelByAsn1(): ParseToSignerInfoModel, len(seq.Bytes):", len(seq.Bytes),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse signer by asn1",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseRoaModelByAsn1():ParseToSignerInfoModel  roaModel.SignerInfoModel:", jsonutil.MarshalJson(roaModel.SignerInfoModel))

		}
	}
	roaModel.RoaIpAddressModels = roaIpAddressModels
	belogs.Info("ParseRoaModelByAsn1(): roaModel:", roaModel.String(), "  time(s):", time.Since(start))
	return
}
