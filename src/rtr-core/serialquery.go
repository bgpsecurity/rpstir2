package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrSerialQueryModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
	SerialNumber    uint32 `json:"serialNumber"`
}

func NewRtrSerialQueryModel(protocolVersion uint8, sessionId uint16,
	serialNumber uint32) *RtrSerialQueryModel {
	return &RtrSerialQueryModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_SERIAL_QUERY,
		SessionId:       sessionId,
		Length:          12,
		SerialNumber:    serialNumber,
	}
}

func (p *RtrSerialQueryModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SerialNumber)
	return wr.Bytes()
}
func (p *RtrSerialQueryModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrSerialQueryModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}
func (p *RtrSerialQueryModel) GetPduType() uint8 {
	return p.PduType
}
func ParseToSerialQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get sessionId fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get SessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 12 {
		belogs.Error("ParseToSerialQuery():PDU_TYPE_SERIAL_QUERY,  length must be 12, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is SERIAL QUERY, length must be 12"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get serialNumber fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}
	sq := NewRtrSerialQueryModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialQuery():get PDU_TYPE_SERIAL_QUERY, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}

// when len(rtrIncrementals)==0, just return endofdata, it is not an error
func AssembleSerialResponses(rtrIncrementals []model.LabRpkiRtrIncremental,
	rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	rtrMoaIncrementals []model.LabRpkiRtrMoaIncremental,
	protocolVersion uint8, sessionId uint16, serialNumber uint32) (rtrPduModels []RtrPduModel, err error) {
	start := time.Now()
	belogs.Info("AssembleSerialResponses(): len(rtrIncrementals):", len(rtrIncrementals),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "  len(rtrMoaIncrementals):", len(rtrMoaIncrementals),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	if protocolVersion != PDU_PROTOCOL_VERSION_0 &&
		protocolVersion != PDU_PROTOCOL_VERSION_1 &&
		protocolVersion != PDU_PROTOCOL_VERSION_2 {
		belogs.Error("AssembleSerialResponses(): protocolVersion is error, protocolVersion:", protocolVersion)
		return nil, errors.New("protocolVersion is error")
	}

	rtrPduModels = make([]RtrPduModel, 0)
	dataAvailable := false

	// start response
	cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
	rtrPduModels = append(rtrPduModels, cacheResponseModel)

	belogs.Debug("AssembleSerialResponses(): cacheResponseModel:", jsonutil.MarshalJson(cacheResponseModel),
		"   protocolVersion:", protocolVersion)

	//rtr incr from roa rtr
	if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
		protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2 {
		if len(rtrIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses(): will get rtrIncrementals, len(rtrIncrementals):", len(rtrIncrementals),
				"   protocolVersion:", protocolVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr incr to response
			rtrIncrementalPduModels, err := convertRtrIncrementalsToRtrPduModels(rtrIncrementals, protocolVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrIncrementalPduModels...)
			dataAvailable = true
			belogs.Info("AssembleSerialResponses(): get rtrIncrementalPduModels, len(rtrIncrementalPduModels):", len(rtrIncrementalPduModels),
				"  time(s):", time.Since(start))
		}
	}

	if protocolVersion == PDU_PROTOCOL_VERSION_2 {
		//rtr incr from asa rtr
		if len(rtrAsaIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses():  will get rtrAsaIncrementals,  len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
				"   protocolVersion:", protocolVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr asa incr to response
			rtrAsaIncrementalPduModels, err := convertRtrAsaIncrementalsToRtrPduModels(rtrAsaIncrementals, protocolVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrAsaIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsaIncrementalPduModels...)
			dataAvailable = true
			belogs.Debug("AssembleSerialResponses():  get rtrAsaIncrementalPduModels, len(rtrAsaIncrementalPduModels):", len(rtrAsaIncrementalPduModels))
		}
		//rtr incr from moa rtr
		if len(rtrMoaIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses():  will get rtrMoaIncrementals,  len(rtrMoaIncrementals):", len(rtrMoaIncrementals),
				"   protocolVersion:", protocolVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr moa incr to response
			rtrMoaIncrementalPduModels, err := convertRtrMoaIncrementalsToRtrPduModels(rtrMoaIncrementals, protocolVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrMoaIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrMoaIncrementalPduModels...)
			dataAvailable = true
			belogs.Debug("AssembleSerialResponses():  get rtrMoaIncrementalPduModels, len(rtrMoaIncrementalPduModels):", len(rtrMoaIncrementalPduModels))
		}

	}

	if !dataAvailable {
		errorReportModel := NewRtrErrorReportModel(protocolVersion, PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)
		rtrPduModels = append(rtrPduModels, errorReportModel)
		belogs.Info("AssembleSerialResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			"  receive protocolVersion:", protocolVersion, "   sessionId:", sessionId,
			"  serialNumber:", serialNumber, "  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	} else {
		// end response
		endOfDataModel := AssembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		rtrPduModels = append(rtrPduModels, endOfDataModel)
		belogs.Debug("AssembleSerialResponses(): endOfDataModel:", jsonutil.MarshalJson(endOfDataModel))

		belogs.Info("AssembleSerialResponses(): will send Cache Response of incrtmental rtr,",
			"   protocolVersion:", protocolVersion,
			"   sessionId:", sessionId, "  serialNumber:", serialNumber,
			"   len(rtrIncrementals):", len(rtrIncrementals),
			"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
			"   len(rtrPduModels):", len(rtrPduModels), "  time(s):", time.Since(start))
	}
	belogs.Debug("AssembleSerialResponses(): ok, rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	return rtrPduModels, nil

}

func convertRtrIncrementalsToRtrPduModels(rtrIncrementals []model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i, _ := range rtrIncrementals {
		rtrPduModel, err := convertRtrIncrementalToRtrPduModel(&rtrIncrementals[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrIncrementalsToRtrPduModels(): convertRtrIncrementalToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrIncrementalsToRtrPduModels(): len(rtrIncrementals):", len(rtrIncrementals), " len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrIncrementalToRtrPduModel(rtrIncremental *model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrIncremental.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := NewRtrIpv4PrefixModel(protocolVersion, getModelFlagsFromStyle(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv4, uint32(rtrIncremental.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := NewRtrIpv6PrefixModel(protocolVersion, getModelFlagsFromStyle(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv6, uint32(rtrIncremental.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}

func convertRtrAsaIncrementalsToRtrPduModels(rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	protocolVersion uint8) (rtrAsaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrAsaIncrementalsToRtrPduModels(): len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "  protocolVersion:", protocolVersion)

	start := time.Now()
	rtrAsaPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsaIncrementals {
		providerAsns := make([]uint32, 0)
		err := jsonutil.UnmarshalJson(rtrAsaIncrementals[i].ProviderAsns, &providerAsns)
		if err != nil {
			belogs.Error("convertRtrAsaIncrementalsToRtrPduModels(): UnmarshalJson ProviderAsns fail,rtrAsaIncrementals[i].ProviderAsns",
				rtrAsaIncrementals[i].ProviderAsns, err)
			return nil, err
		}

		rtrPduModel := NewRtrAsaModel(protocolVersion, getModelFlagsFromStyle(rtrAsaIncrementals[i].Style),
			uint32(rtrAsaIncrementals[i].CustomerAsn), providerAsns)
		rtrAsaPduModels = append(rtrAsaPduModels, rtrPduModel)
		belogs.Debug("convertRtrAsaIncrementalsToRtrPduModels(): rtrPduModel: ", jsonutil.MarshalJson(rtrPduModel))
	}
	belogs.Info("convertRtrAsaIncrementalsToRtrPduModels(): len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
		" len(rtrAsaPduModels):", len(rtrAsaPduModels), "  time(s):", time.Since(start))

	return rtrAsaPduModels, nil
}

func convertRtrMoaIncrementalsToRtrPduModels(rtrMoaIncrementals []model.LabRpkiRtrMoaIncremental,
	protocolVersion uint8) (rtrMoaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrMoaIncrementalsToRtrPduModels(): len(rtrMoaIncrementals):", len(rtrMoaIncrementals), "  protocolVersion:", protocolVersion)

	start := time.Now()
	rtrMoaPduModels = make([]RtrPduModel, 0)
	for i := range rtrMoaIncrementals {
		rtrFormatV6Bytes, prefixLength, err := convertAddressPrefixToRtrFormatBytes(rtrMoaIncrementals[i].Ipv6MappingPrefix)
		if err != nil {
			belogs.Error("convertRtrMoaIncrementalsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes fail,rtrMoaIncrementals[i].Ipv6MappingPrefix",
				rtrMoaIncrementals[i].Ipv6MappingPrefix, err)
			return nil, err
		}
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], rtrFormatV6Bytes[:])
		rtrPduModel := NewRtrMoaModelFromDb(protocolVersion, getModelFlagsFromStyle(rtrMoaIncrementals[i].Style),
			uint8(prefixLength), ipv6)
		belogs.Debug("convertRtrMoaIncrementalsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes ipv6 ok",
			"  rtrMoaIncrementals[i].Ipv6MappingPrefix", rtrMoaIncrementals[i].Ipv6MappingPrefix,
			"rtrFormatV6Bytes", convert.PrintBytesOneLine(rtrFormatV6Bytes),
			"ipv6", ipv6, "prefixLength", prefixLength)

		ipv4Prefixs := make([]string, 0)
		err = jsonutil.UnmarshalJson(rtrMoaIncrementals[i].Ipv4Prefixes, &ipv4Prefixs)
		if err != nil {
			belogs.Error("convertRtrMoaIncrementalsToRtrPduModels(): UnmarshalJson ProviderAsns fail, rtrMoaIncrementals[i].Ipv4Prefixes",
				rtrMoaIncrementals[i].Ipv4Prefixes, err)
			return nil, err
		}
		for _, ipv4Prefix := range ipv4Prefixs {
			rtrFormatV4Bytes, prefixLength, err := convertAddressPrefixToRtrFormatBytes(ipv4Prefix)
			if err != nil {
				belogs.Error("convertRtrMoaIncrementalsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes fail,ipv4Prefix",
					ipv4Prefix, err)
				return nil, err
			}
			ipv4 := [4]byte{0x00}
			copy(ipv4[:], rtrFormatV4Bytes[:])
			ipv4PrefixEntry := IPv4Prefix{
				IPv4PrefixLength: uint8(prefixLength),
				IPv4Prefix:       ipv4,
			}
			belogs.Debug("convertRtrMoaIncrementalsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes ipv4 ok",
				" ipv4Prefix:", ipv4Prefix, "  ipv4PrefixEntry", ipv4PrefixEntry, "rtrFormatV4Bytes", convert.PrintBytesOneLine(rtrFormatV4Bytes),
				"ipv4", ipv4, "prefixLength", prefixLength)

			rtrPduModel.AddIPv4Prefix(ipv4PrefixEntry)
		}
		rtrMoaPduModels = append(rtrMoaPduModels, rtrPduModel)
	}
	belogs.Info("convertRtrMoaIncrementalsToRtrPduModels(): len(rtrMoaIncrementals):", len(rtrMoaIncrementals),
		" len(rtrMoaPduModels):", len(rtrMoaPduModels), "  time(s):", time.Since(start))

	return rtrMoaPduModels, nil
}
