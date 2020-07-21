package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"

	rtrmodel "rtr/model"
)

func ParseToRouterKey(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	/*
		ProtocolVersion      uint8    `json:"protocolVersion"`
		PduType              uint8    `json:"pduType"`
		Flags                uint8    `json:"flags"`
		Zero                 uint8    `json:"zero"`
		Length               uint32   `json:"length"`
		SubjectKeyIdentifier [20]byte `json:"subjectKeyIdentifier"`
		Asn                  uint32   `json:"asn"`
		SubjectPublicKeyInfo uint32   `json:"subjectPublicKeyInfo"`
	*/
	var flags uint8
	var zero uint8
	var length uint32
	var subjectKeyIdentifier [20]byte
	var asn uint32
	var subjectPublicKeyInfo uint32

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToRouterKey(): PDU_TYPE_ROUTER_KEY get flags fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	if flags != 0 && flags != 1 {
		belogs.Error("ParseToRouterKey():PDU_TYPE_ROUTER_KEY, flags must be 0 or 1, ", buf, flags)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is ROUTER KEY, flags must be 0 or 1"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get zero
	err = binary.Read(buf, binary.BigEndian, &zero)
	if err != nil {
		belogs.Error("ParseToRouterKey(): PDU_TYPE_ROUTER_KEY get zero fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToRouterKey(): PDU_TYPE_ROUTER_KEY get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get subjectKeyIdentifier
	err = binary.Read(buf, binary.BigEndian, &subjectKeyIdentifier)
	if err != nil {
		belogs.Error("ParseToRouterKey(): PDU_TYPE_ROUTER_KEY get subjectKeyIdentifier fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get subjectKeyIdentifier")
		return rtrPduModel, rtrError
	}

	// get asn
	err = binary.Read(buf, binary.BigEndian, &asn)
	if err != nil {
		belogs.Error("ParseToRouterKey(): PDU_TYPE_ROUTER_KEY get asn fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asn")
		return rtrPduModel, rtrError
	}

	// get subjectPublicKeyInfo
	err = binary.Read(buf, binary.BigEndian, &subjectPublicKeyInfo)
	if err != nil {
		belogs.Error("ParseToRouterKey(): PDU_TYPE_ROUTER_KEY get subjectPublicKeyInfo fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get subjectPublicKeyInfo")
		return rtrPduModel, rtrError
	}

	sq := rtrmodel.NewRtrRouterKeyModel(flags, subjectKeyIdentifier,
		asn, subjectPublicKeyInfo)

	belogs.Debug("ParseToRouterKey():get PDU_TYPE_ROUTER_KEY ", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
