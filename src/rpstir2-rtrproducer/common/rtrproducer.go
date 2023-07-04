package common

import (
	"errors"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/guregu/null"
	model "rpstir2-model"
)

func ConvertSlurmAddressFamilyToRtr(addressFamilyStr string) (addressFamilyIpv4 null.Int, addressFamilyIpv6 null.Int, err error) {
	if addressFamilyStr == "" {
		addressFamilyIpv4 = null.IntFrom(0)
		addressFamilyIpv6 = null.IntFrom(1)
	} else if strings.EqualFold(addressFamilyStr, model.SLURM_PROVIDER_ASNS_ADDRESS_FAMILY_IPV4) {
		addressFamilyIpv4 = null.IntFrom(0)
		addressFamilyIpv6 = null.NewInt(0, false)
	} else if strings.EqualFold(addressFamilyStr, model.SLURM_PROVIDER_ASNS_ADDRESS_FAMILY_IPV6) {
		addressFamilyIpv4 = null.NewInt(0, false)
		addressFamilyIpv6 = null.IntFrom(1)
	} else {
		belogs.Error("ConvertSlurmAddressFamilyToRtr(): addressFamilyStr type fail, addressFamilyStr:", addressFamilyStr)
		return null.NewInt(0, false), null.NewInt(0, false), errors.New("addressFamily is error")
	}
	belogs.Debug("ConvertSlurmAddressFamilyToRtr(): addressFamilyStr:", addressFamilyStr, "  addressFamilyIpv4:", addressFamilyIpv4,
		"   addressFamilyIpv6:", addressFamilyIpv6)
	return addressFamilyIpv4, addressFamilyIpv6, nil
}

func ConvertAsaAddressFamilyToRtr(addressFamilyInt null.Int) (addressFamilyIpv4 null.Int, addressFamilyIpv6 null.Int, err error) {
	if !addressFamilyInt.Valid {
		addressFamilyIpv4 = null.IntFrom(0)
		addressFamilyIpv6 = null.IntFrom(1)
	} else if addressFamilyInt.ValueOrZero() == 1 {
		addressFamilyIpv4 = null.IntFrom(0)
		addressFamilyIpv6 = null.NewInt(0, false)
	} else if addressFamilyInt.ValueOrZero() == 2 {
		addressFamilyIpv4 = null.NewInt(0, false)
		addressFamilyIpv6 = null.IntFrom(1)
	} else {
		belogs.Error("ConvertAsaAddressFamilyToRtr(): addressFamilyInt type fail, addressFamilyInt:", addressFamilyInt)
		return null.NewInt(0, false), null.NewInt(0, false), errors.New("addressFamily is error")
	}
	belogs.Debug("ConvertAsaAddressFamilyToRtr(): addressFamilyInt:", addressFamilyInt, "  addressFamilyIpv4:", addressFamilyIpv4,
		"   addressFamilyIpv6:", addressFamilyIpv6)
	return addressFamilyIpv4, addressFamilyIpv6, nil
}
