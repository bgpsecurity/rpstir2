package rtrserver

import (
	"bytes"
	"encoding/binary"

	"github.com/cpusoft/goutil/convert"
)

const (
	PROTOCOL_VERSION_0 = 0
	PROTOCOL_VERSION_1 = 1

	PDU_TYPE_SERIAL_NOTIFY  = 0
	PDU_TYPE_SERIAL_QUERY   = 1
	PDU_TYPE_RESET_QUERY    = 2
	PDU_TYPE_CACHE_RESPONSE = 3
	PDU_TYPE_IPV4_PREFIX    = 4
	PDU_TYPE_IPV6_PREFIX    = 6
	PDU_TYPE_END_OF_DATA    = 7
	PDU_TYPE_CACHE_RESET    = 8
	//PDU_TYPE_RESERVED       = 9
	PDU_TYPE_ROUTER_KEY   = 9
	PDU_TYPE_ERROR_REPORT = 10

	// min pdu type length is reset query
	PDU_TYPE_MIN_LEN = 8

	// error code
	PDU_TYPE_ERROR_CODE_CORRUPT_DATA                    = 0
	PDU_TYPE_ERROR_CODE_INTERNAL_ERROR                  = 1
	PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE               = 2
	PDU_TYPE_ERROR_CODE_INVALID_REQUEST                 = 3
	PDU_TYPE_ERROR_CODE_UNSUPPORTED_PROTOCOL_VERSION    = 4
	PDU_TYPE_ERROR_CODE_UNSUPPORTED_PDU_TYPE            = 5
	PDU_TYPE_ERROR_CODE_WITHDRAWAL_OF_UNKNOWN_RECORD    = 6
	PDU_TYPE_ERROR_CODE_DUPLICATE_ANNOUNCEMENT_RECEIVED = 7
	PDU_TYPE_ERROR_CODE_UNEXPECTED_PROTOCOL_VERSION     = 8

	// seconds.
	PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_MIN         = 1
	PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_MAX         = 86400
	PDU_TYPE_END_OF_DATA_REFRESH_INTERVAL_RECOMMENDED = 3600

	PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_MIN         = 1
	PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_MAX         = 7200
	PDU_TYPE_END_OF_DATA_RETRY_INTERVAL_RECOMMENDED = 600

	PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_MIN         = 600
	PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_MAX         = 172800
	PDU_TYPE_END_OF_DATA_EXPIRE_INTERVAL_RECOMMENDED = 7200

	UINT32_MAX = ^uint32(0)
)

type RtrPduModel interface {
	Bytes() []byte
	PrintBytes() string
	GetProtocolVersion() uint8
	GetPduType() uint8
}

type RtrSerialNotifyModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
	SerialNumber    uint32 `json:"serialNumber"`
}

func NewRtrSerialNotifyModel(protocolVersion uint8, sessionId uint16, serialNumber uint32) *RtrSerialNotifyModel {
	return &RtrSerialNotifyModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_SERIAL_NOTIFY,
		SessionId:       sessionId,
		Length:          12,
		SerialNumber:    serialNumber,
	}
}

func (p *RtrSerialNotifyModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SerialNumber)
	return wr.Bytes()
}
func (p *RtrSerialNotifyModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrSerialNotifyModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}
func (p *RtrSerialNotifyModel) GetPduType() uint8 {
	return p.PduType
}

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

type RtrCacheResponseModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
}

func NewRtrCacheResponseModel(protocolVersion uint8, sessionId uint16) *RtrCacheResponseModel {
	return &RtrCacheResponseModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_CACHE_RESPONSE,
		SessionId:       sessionId,
		Length:          8,
	}
}

func (p *RtrCacheResponseModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)
	binary.Write(wr, binary.BigEndian, p.Length)
	return wr.Bytes()
}

func (p *RtrCacheResponseModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrCacheResponseModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrCacheResponseModel) GetPduType() uint8 {
	return p.PduType
}

type RtrIpv4PrefixModel struct {
	ProtocolVersion uint8   `json:"protocolVersion"`
	PduType         uint8   `json:"pduType"`
	Zero0           uint16  `json:"zero0"`
	Length          uint32  `json:"length"`
	Flags           uint8   `json:"flags"`
	PrefixLength    uint8   `json:"prefixLength"`
	MaxLength       uint8   `json:"maxLength"`
	Zero1           uint8   `json:"zero1"`
	Ipv4Prefix      [4]byte `json:"ipv4Prefix"`
	Asn             uint32  `json:"asn"`
}

func NewRtrIpv4PrefixModel(protocolVersion uint8, flags uint8,
	prefixLength uint8, maxLength uint8, ipv4Prefix [4]byte, asn uint32) *RtrIpv4PrefixModel {
	return &RtrIpv4PrefixModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_IPV4_PREFIX,
		Zero0:           0,
		Length:          20,
		Flags:           flags,
		PrefixLength:    prefixLength,
		MaxLength:       maxLength,
		Zero1:           0,
		Ipv4Prefix:      ipv4Prefix,
		Asn:             asn,
	}
}

func (p *RtrIpv4PrefixModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.PrefixLength)
	binary.Write(wr, binary.BigEndian, p.MaxLength)
	binary.Write(wr, binary.BigEndian, p.Zero1)
	binary.Write(wr, binary.BigEndian, p.Ipv4Prefix)
	binary.Write(wr, binary.BigEndian, p.Asn)
	return wr.Bytes()
}
func (p *RtrIpv4PrefixModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrIpv4PrefixModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrIpv4PrefixModel) GetPduType() uint8 {
	return p.PduType
}

type RtrIpv6PrefixModel struct {
	ProtocolVersion uint8    `json:"protocolVersion"`
	PduType         uint8    `json:"pduType"`
	Zero0           uint16   `json:"zero0"`
	Length          uint32   `json:"length"`
	Flags           uint8    `json:"flags"`
	PrefixLength    uint8    `json:"prefixLength"`
	MaxLength       uint8    `json:"maxLength"`
	Zero1           uint8    `json:"zero1"`
	Ipv6Prefix      [16]byte `json:"ipv6Prefix"`
	Asn             uint32   `json:"asn"`
}

func NewRtrIpv6PrefixModel(protocolVersion uint8, flags uint8,
	prefixLength uint8, maxLength uint8, ipv6Prefix [16]byte, asn uint32) *RtrIpv6PrefixModel {
	return &RtrIpv6PrefixModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_IPV6_PREFIX,
		Zero0:           0,
		Length:          32,
		Zero1:           0,
		Flags:           flags,
		PrefixLength:    prefixLength,
		MaxLength:       maxLength,
		Ipv6Prefix:      ipv6Prefix,
		Asn:             asn,
	}
}

func (p *RtrIpv6PrefixModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.PrefixLength)
	binary.Write(wr, binary.BigEndian, p.MaxLength)
	binary.Write(wr, binary.BigEndian, p.Zero1)
	binary.Write(wr, binary.BigEndian, p.Ipv6Prefix)
	binary.Write(wr, binary.BigEndian, p.Asn)
	return wr.Bytes()
}
func (p *RtrIpv6PrefixModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrIpv6PrefixModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrIpv6PrefixModel) GetPduType() uint8 {
	return p.PduType
}

type RtrEndOfDataModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
	SerialNumber    uint32 `json:"serialNumber"`
	RefreshInterval uint32 `json:"refreshInterval"`
	RetryInterval   uint32 `json:"retryInterval"`
	ExpireInterval  uint32 `json:"expireInterval"`
}

func NewRtrEndOfDataModel(protocolVersion uint8, sessionId uint16,
	serialNumber uint32, refreshInterval uint32,
	retryInterval uint32, expireInterval uint32) *RtrEndOfDataModel {
	if protocolVersion == PROTOCOL_VERSION_0 {
		return &RtrEndOfDataModel{
			ProtocolVersion: protocolVersion,
			PduType:         PDU_TYPE_END_OF_DATA,
			SessionId:       sessionId,
			Length:          12,
			SerialNumber:    serialNumber,
		}

	} else if protocolVersion == PROTOCOL_VERSION_1 {
		return &RtrEndOfDataModel{
			ProtocolVersion: protocolVersion,
			PduType:         PDU_TYPE_END_OF_DATA,
			SessionId:       sessionId,
			Length:          24,
			SerialNumber:    serialNumber,
			RefreshInterval: refreshInterval,
			RetryInterval:   retryInterval,
			ExpireInterval:  expireInterval,
		}
	}
	return nil

}
func (p *RtrEndOfDataModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)

	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SerialNumber)
	if p.ProtocolVersion == PROTOCOL_VERSION_1 {
		binary.Write(wr, binary.BigEndian, p.RefreshInterval)
		binary.Write(wr, binary.BigEndian, p.RetryInterval)
		binary.Write(wr, binary.BigEndian, p.ExpireInterval)
	}

	return wr.Bytes()
}

func (p *RtrEndOfDataModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrEndOfDataModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrEndOfDataModel) GetPduType() uint8 {
	return p.PduType
}

type RtrCacheResetModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	Zero            uint16 `json:"zero"`
	Length          uint32 `json:"length"`
}

func NewRtrCacheResetModel(protocolVersion uint8) *RtrCacheResetModel {
	return &RtrCacheResetModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_CACHE_RESET,
		Zero:            0,
		Length:          8,
	}
}

func (p *RtrCacheResetModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	return wr.Bytes()
}

func (p *RtrCacheResetModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrCacheResetModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrCacheResetModel) GetPduType() uint8 {
	return p.PduType
}

type RtrRouterKeyModel struct {
	ProtocolVersion      uint8    `json:"protocolVersion"`
	PduType              uint8    `json:"pduType"`
	Flags                uint8    `json:"flags"`
	Zero                 uint8    `json:"zero"`
	Length               uint32   `json:"length"`
	SubjectKeyIdentifier [20]byte `json:"subjectKeyIdentifier"`
	Asn                  uint32   `json:"asn"`
	SubjectPublicKeyInfo uint32   `json:"subjectPublicKeyInfo"`
}

func NewRtrRouterKeyModel(flags uint8, subjectKeyIdentifier [20]byte,
	asn uint32, subjectPublicKeyInfo uint32) *RtrRouterKeyModel {
	return &RtrRouterKeyModel{
		ProtocolVersion:      PROTOCOL_VERSION_1,
		PduType:              PDU_TYPE_ROUTER_KEY,
		Flags:                flags,
		Zero:                 0,
		SubjectKeyIdentifier: subjectKeyIdentifier,
		Asn:                  asn,
		SubjectPublicKeyInfo: subjectPublicKeyInfo,
	}
}

func (p *RtrRouterKeyModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SubjectKeyIdentifier)
	binary.Write(wr, binary.BigEndian, p.Asn)
	binary.Write(wr, binary.BigEndian, p.SubjectPublicKeyInfo)
	return wr.Bytes()
}

func (p *RtrRouterKeyModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrRouterKeyModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrRouterKeyModel) GetPduType() uint8 {
	return p.PduType
}

type RtrErrorReportModel struct {
	ProtocolVersion        uint8  `json:"protocolVersion"`
	PduType                uint8  `json:"pduType"`
	ErrorCode              uint16 `json:"errorCode"`
	Length                 uint32 `json:"length"`
	LengthOfEncapsulated   uint32 `json:"lengthOfEncapsulated"`
	ErroneousPdu           []byte `json:"erroneousPdu"`
	LengthOfErrorText      uint32 `json:"lengthOfErrorText"`
	ErrorDiagnosticMessage []byte `json:"errorDiagnosticMessage"`
}

// erroneousPdu and errorDiagnosticMessage can be nil
func NewRtrErrorReportModel(protocolVersion uint8, errorCode uint16,
	erroneousPdu []byte, errorDiagnosticMessage []byte) *RtrErrorReportModel {
	erm := &RtrErrorReportModel{PduType: PDU_TYPE_ERROR_REPORT}
	erm.ProtocolVersion = protocolVersion
	erm.ErrorCode = errorCode
	erm.LengthOfEncapsulated = uint32(len(erroneousPdu))
	erm.ErroneousPdu = erroneousPdu
	erm.LengthOfErrorText = uint32(len(errorDiagnosticMessage))
	erm.ErrorDiagnosticMessage = errorDiagnosticMessage
	// (protocolversion+pdutype+errorCode)+length + lengthofencapsulatedpdu + ErroneousPDU + LengthOfErrorText + errorDiagnosticMessage
	erm.Length = 4 + 4 + 4 + uint32(len(erroneousPdu)) + 4 + uint32(len(errorDiagnosticMessage))

	return erm
}

// erroneousPdu and errorDiagnosticMessage can be nil
func NewRtrErrorReportModelByRtrError(rtrError *RtrError) *RtrErrorReportModel {

	return NewRtrErrorReportModel(rtrError.ProtocolVersion, rtrError.ErrorCode,
		rtrError.ErroneousPdu, rtrError.ErrorDiagnosticMessage)
}

func (p *RtrErrorReportModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.ErrorCode)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.LengthOfEncapsulated)
	if len(p.ErroneousPdu) > 0 {
		binary.Write(wr, binary.BigEndian, p.ErroneousPdu)
	}
	binary.Write(wr, binary.BigEndian, p.LengthOfErrorText)
	if len(p.ErrorDiagnosticMessage) > 0 {
		binary.Write(wr, binary.BigEndian, p.ErrorDiagnosticMessage)
	}
	return wr.Bytes()
}

func (p *RtrErrorReportModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrErrorReportModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrErrorReportModel) GetPduType() uint8 {
	return p.PduType
}

// withdraw-->0, announce-->1
func GetIpPrefixModelFlags(style string) uint8 {
	if style == "withdraw" {
		return 0
	} else if style == "announce" {
		return 1
	}
	return 0
}

type RtrError struct {
	Err error `json:"err"`
	// if get error pdu ,do not send response
	NeedSendResponse bool `json:"needSendResponse"`

	ProtocolVersion        uint8  `json:"protocolVersion"`
	ErrorCode              uint16 `json:"errorCode"`
	ErroneousPdu           []byte `json:"erroneousPdu"`
	ErrorDiagnosticMessage []byte `json:"errorDiagnosticMessage"`
}

func NewRtrError(err error, needSendResponse bool, protocolVersion uint8, errorCode uint16,
	buf *bytes.Reader, errorDiagnosticMessage string) *RtrError {
	var erroneousPdu []byte
	if buf != nil {
		buf.Seek(0, 0)
		erroneousPdu = make([]byte, buf.Size())
		buf.Read(erroneousPdu)
	} else {
		erroneousPdu = nil
	}

	rtrError := &RtrError{
		Err:                    err,
		NeedSendResponse:       needSendResponse,
		ProtocolVersion:        protocolVersion,
		ErrorCode:              errorCode,
		ErroneousPdu:           erroneousPdu,
		ErrorDiagnosticMessage: []byte(errorDiagnosticMessage),
	}

	return rtrError
}

func (p *RtrError) Error() string {
	return p.Err.Error()
}
func (p *RtrError) Unwrap() error {
	return p.Err
}
