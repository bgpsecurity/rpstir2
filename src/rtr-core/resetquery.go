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

type RtrResetQueryModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	Zero            uint16 `json:"zero"`
	Length          uint32 `json:"length"`
}

func NewRtrResetQueryModel(protocolVersion uint8) *RtrResetQueryModel {
	return &RtrResetQueryModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_RESET_QUERY,
		Zero:            0,
		Length:          8,
	}
}

func (p *RtrResetQueryModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	return wr.Bytes()
}
func (p *RtrResetQueryModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrResetQueryModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}
func (p *RtrResetQueryModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToResetQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var zero16 uint16
	var length uint32

	// get zero16
	err = binary.Read(buf, binary.BigEndian, &zero16)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get zero fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToResetQuery():PDU_TYPE_RESET_QUERY, length must be 8, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is RESET QUERY, length must be 8"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	rq := NewRtrResetQueryModel(protocolVersion)
	belogs.Debug("ParseToResetQuery():get PDU_TYPE_RESET_QUERY, buf:", buf, "   rq:", jsonutil.MarshalJson(rq))
	return rq, nil
}

// when len(rtrFull)==0, it is an error with no_data_available
func AssembleResetResponses(rtrFulls []model.LabRpkiRtrFull,
	rtrAsaFulls []model.LabRpkiRtrAsaFull, rtrMoaFulls []model.LabRpkiRtrMoaFull,
	protocolVersion uint8, sessionId uint16, serialNumber uint32) (rtrPduModels []RtrPduModel, err error) {
	start := time.Now()
	belogs.Info("AssembleResetResponses(): len(rtrFulls):", len(rtrFulls),
		"   len(rtrAsaFulls):", len(rtrAsaFulls), " len(rtrMoaFulls):", len(rtrMoaFulls),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	if protocolVersion != PDU_PROTOCOL_VERSION_0 &&
		protocolVersion != PDU_PROTOCOL_VERSION_1 &&
		protocolVersion != PDU_PROTOCOL_VERSION_2 {
		belogs.Error("AssembleResetResponses(): protocolVersion is error, protocolVersion:", protocolVersion)
		return nil, errors.New("protocolVersion is error")
	}

	rtrPduModels = make([]RtrPduModel, 0)
	dataAvailable := false
	// start response
	cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
	rtrPduModels = append(rtrPduModels, cacheResponseModel)

	belogs.Debug("AssembleResetResponses(): cacheResponseModel:", jsonutil.MarshalJson(cacheResponseModel),
		"   protocolVersion:", protocolVersion)
	// rtr full from roa rtr
	if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
		protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2 {
		if len(rtrFulls) > 0 {

			belogs.Debug("AssembleResetResponses(): will get rtrFullPduModels, len(rtrFulls)>0, len(rtrFulls):", len(rtrFulls),
				"   protocolVersion:", protocolVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr full to response
			rtrFullPduModels, err := convertRtrFullsToRtrPduModels(rtrFulls, protocolVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrFullPduModels, len(rtrFullPduModels):", len(rtrFullPduModels),
				"  time(s):", time.Since(start))
		}
	}
	if protocolVersion == PDU_PROTOCOL_VERSION_2 {
		//rtr full from asa rtr
		if len(rtrAsaFulls) > 0 {
			belogs.Debug("AssembleResetResponses(): will get rtrAsaFullPduModels, len(rtrAsaFulls):", len(rtrAsaFulls),
				"   protocolVersion:", protocolVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr asa full to response
			rtrAsaFullPduModels, err := convertRtrAsaFullsToRtrPduModels(rtrAsaFulls, protocolVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrAsaFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsaFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrAsaFullPduModels, len(rtrAsaFullPduModels):", len(rtrAsaFullPduModels),
				"  time(s):", time.Since(start))
		}

		//rtr full from moa rtr
		if len(rtrMoaFulls) > 0 {
			belogs.Debug("AssembleResetResponses(): will get rtrMoaFullPduModels, len(rtrMoaFulls):", len(rtrMoaFulls),
				"   protocolVersion:", protocolVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr moa full to response
			rtrMoaFullPduModels, err := convertRtrMoaFullsToRtrPduModels(rtrMoaFulls, protocolVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrMoaFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrMoaFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrMoaFullPduModels, len(rtrMoaFullPduModels):", len(rtrMoaFullPduModels),
				"  time(s):", time.Since(start))
		}
	}

	if !dataAvailable {
		errorReportModel := NewRtrErrorReportModel(protocolVersion, PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)
		rtrPduModels = append(rtrPduModels, errorReportModel)
		belogs.Info("AssembleResetResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			"  receive protocolVersion:", protocolVersion, "   sessionId:", sessionId,
			"  serialNumber:", serialNumber, "  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	} else {
		// end response
		endOfDataModel := AssembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		rtrPduModels = append(rtrPduModels, endOfDataModel)
		belogs.Debug("AssembleResetResponses(): endOfDataModel:", jsonutil.MarshalJson(endOfDataModel))

		belogs.Info("AssembleResetResponses(): will send Cache Response of all rtr,",
			"   protocolVersion:", protocolVersion,
			"   sessionId:", sessionId, "  serialNumber:", serialNumber,
			"   len(rtrFulls):", len(rtrFulls),
			"   len(rtrAsaFulls):", len(rtrAsaFulls),
			"   len(rtrPduModels):", len(rtrPduModels), "  time(s):", time.Since(start))
	}
	belogs.Debug("AssembleResetResponses(): ok, rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	return rtrPduModels, nil

}

func convertRtrFullToRtrPduModel(rtrFull *model.LabRpkiRtrFull, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	belogs.Debug("convertRtrFullToRtrPduModel(): input param", "rtrFull.Address", rtrFull.Address,
		"protocolVersion", protocolVersion)
	if len(rtrFull.Address) == 0 {
		belogs.Error("convertRtrFullToRtrPduModel(): rtrFull.Address is empty fail",
			"rtrFull.Address", rtrFull.Address)
		return rtrPduModel, errors.New("rtrFull address is empty")
	}

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrFull.Address)
	if err != nil {
		belogs.Error("convertRtrFullToRtrPduModel(): AddressToRtrFormatByte fail", "rtrFull.Address", rtrFull.Address,
			"err", err)
		return rtrPduModel, err
	}
	belogs.Debug("convertRtrFullToRtrPduModel(): AddressToRtrFormatByte ok", "rtrFull.Address", rtrFull.Address,
		"ipHex", ipHex, "ipType", ipType)

	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := NewRtrIpv4PrefixModel(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv4, uint32(rtrFull.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := NewRtrIpv6PrefixModel(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv6, uint32(rtrFull.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}

func convertRtrFullsToRtrPduModels(rtrFulls []model.LabRpkiRtrFull,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrFulls {
		rtrPduModel, err := convertRtrFullToRtrPduModel(&rtrFulls[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrFullsToRtrPduModels(): convertRtrFullToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrFullsToRtrPduModels(): len(rtrFulls):", len(rtrFulls), " len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrAsaFullsToRtrPduModels(rtrAsaFulls []model.LabRpkiRtrAsaFull,
	protocolVersion uint8) (rtrAsaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrAsaFullsToRtrPduModels(): len(rtrAsaFulls):", len(rtrAsaFulls), "  protocolVersion:", protocolVersion)

	start := time.Now()
	rtrAsaPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsaFulls {

		providerAsns := make([]uint32, 0)
		err := jsonutil.UnmarshalJson(rtrAsaFulls[i].ProviderAsns, &providerAsns)
		if err != nil {
			belogs.Error("convertRtrAsaFullsToRtrPduModels(): UnmarshalJson ProviderAsns fail,rtrAsaFulls[i].ProviderAsns",
				rtrAsaFulls[i].ProviderAsns, err)
			return nil, err
		}
		rtrPduModel := NewRtrAsaModel(protocolVersion, PDU_FLAG_ANNOUNCE,
			uint32(rtrAsaFulls[i].CustomerAsn), providerAsns)

		rtrAsaPduModels = append(rtrAsaPduModels, rtrPduModel)
		belogs.Debug("convertRtrAsaFullsToRtrPduModels(): rtrPduModel: ", jsonutil.MarshalJson(rtrPduModel))
	}
	belogs.Info("convertRtrAsaFullsToRtrPduModels(): len(rtrAsaFulls):", len(rtrAsaFulls),
		" len(rtrAsaPduModels):", len(rtrAsaPduModels), "  time(s):", time.Since(start))
	return rtrAsaPduModels, nil
}

func convertRtrMoaFullsToRtrPduModels(rtrMoaFulls []model.LabRpkiRtrMoaFull,
	protocolVersion uint8) (rtrMoaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrMoaFullsToRtrPduModels(): len(rtrMoaFulls):", len(rtrMoaFulls), "  protocolVersion:", protocolVersion)

	start := time.Now()
	rtrMoaPduModels = make([]RtrPduModel, 0)
	for i := range rtrMoaFulls {

		rtrFormatV6Bytes, prefixLength, err := convertAddressPrefixToRtrFormatBytes(rtrMoaFulls[i].Ipv6MappingPrefix)
		if err != nil {
			belogs.Error("convertRtrMoaFullsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes fail,rtrMoaFulls[i].Ipv6MappingPrefix",
				rtrMoaFulls[i].Ipv6MappingPrefix, err)
			return nil, err
		}
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], rtrFormatV6Bytes[:])
		rtrPduModel := NewRtrMoaModelFromDb(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(prefixLength), ipv6)
		belogs.Debug("convertRtrMoaFullsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes ipv6 ok",
			"  rtrMoaFulls[i].Ipv6MappingPrefix", rtrMoaFulls[i].Ipv6MappingPrefix,
			"rtrFormatV6Bytes", convert.PrintBytesOneLine(rtrFormatV6Bytes),
			"ipv6", ipv6, "prefixLength", prefixLength)

		ipv4Prefixs := make([]string, 0)
		err = jsonutil.UnmarshalJson(rtrMoaFulls[i].Ipv4Prefixes, &ipv4Prefixs)
		if err != nil {
			belogs.Error("convertRtrMoaFullsToRtrPduModels(): UnmarshalJson ProviderAsns fail, rtrMoaFulls[i].Ipv4Prefixes",
				rtrMoaFulls[i].Ipv4Prefixes, err)
			return nil, err
		}
		for _, ipv4Prefix := range ipv4Prefixs {
			rtrFormatV4Bytes, prefixLength, err := convertAddressPrefixToRtrFormatBytes(ipv4Prefix)
			if err != nil {
				belogs.Error("convertRtrMoaFullsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes fail,ipv4Prefix",
					ipv4Prefix, err)
				return nil, err
			}
			ipv4 := [4]byte{0x00}
			copy(ipv4[:], rtrFormatV4Bytes[:])
			ipv4PrefixEntry := IPv4Prefix{
				IPv4PrefixLength: uint8(prefixLength),
				IPv4Prefix:       ipv4,
			}
			belogs.Debug("convertRtrMoaFullsToRtrPduModels(): convertAddressPrefixToRtrFormatBytes ipv4 ok",
				"  ipv4Prefix", ipv4Prefix, "  ipv4PrefixEntry:", ipv4PrefixEntry, "rtrFormatV4Bytes", convert.PrintBytesOneLine(rtrFormatV4Bytes),
				"ipv4", ipv4, "prefixLength", prefixLength)

			rtrPduModel.AddIPv4Prefix(ipv4PrefixEntry)
		}

		rtrMoaPduModels = append(rtrMoaPduModels, rtrPduModel)
		belogs.Debug("convertRtrMoaFullsToRtrPduModels(): rtrPduModel: ", jsonutil.MarshalJson(rtrPduModel))
	}
	belogs.Info("convertRtrMoaFullsToRtrPduModels(): len(rtrMoaFulls):", len(rtrMoaFulls),
		" len(rtrMoaPduModels):", len(rtrMoaPduModels), "  time(s):", time.Since(start))
	return rtrMoaPduModels, nil
}
