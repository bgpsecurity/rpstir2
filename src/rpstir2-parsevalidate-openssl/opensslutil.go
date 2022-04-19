package openssl

import (
	"errors"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

func convertAsProviderAttestationToCustomerAsns(asProviderAttestation AsProviderAttestation) (customerAsns []model.CustomerAsn, err error) {
	belogs.Debug("convertAsProviderAttestationToCustomerAsns(): asProviderAttestation:", jsonutil.MarshalJson(asProviderAttestation))

	customerAsn := model.CustomerAsn{}
	customerAsns = make([]model.CustomerAsn, 0)
	if len(asProviderAttestation.AddressFamilyIdentifier) > 0 {
		belogs.Debug("convertAsProviderAttestationToCustomerAsns():AddressFamilyIdentifier:", asProviderAttestation.AddressFamilyIdentifier)
		afi := convert.BytesToBigInt(asProviderAttestation.AddressFamilyIdentifier)
		if afi == nil {
			belogs.Error("convertAsProviderAttestationToCustomerAsns():AddressFamilyIdentifier is not 0x01 or 0x02:",
				asProviderAttestation.AddressFamilyIdentifier)
			return nil, errors.New("AddressFamilyIdentifier is not 0x01 or 0x02")
		}
		addressFamily := null.IntFrom(afi.Int64())
		belogs.Debug("convertAsProviderAttestationToCustomerAsns(): addressFamily:", addressFamily)
		customerAsn.AddressFamily = addressFamily
	}
	customerAsn.CustomerAsn = uint64(asProviderAttestation.CustomerAsn)

	providerAsns := make([]model.ProviderAsn, 0)
	for i := range asProviderAttestation.ProviderAsns {
		var providerAsn model.ProviderAsn
		if len(asProviderAttestation.ProviderAsns[i].AddressFamilyIdentifier) > 0 {
			belogs.Debug("convertAsProviderAttestationToCustomerAsns():(asProviderAttestation.ProviderAsns[i].AddressFamilyIdentifier:",
				asProviderAttestation.ProviderAsns[i].AddressFamilyIdentifier)
			afi := convert.BytesToBigInt(asProviderAttestation.ProviderAsns[i].AddressFamilyIdentifier)
			if afi == nil {
				belogs.Error("convertAsProviderAttestationToCustomerAsns():asProviderAttestation.ProviderAsns[i].AddressFamilyIdentifier is not 0x01 or 0x02:",
					asProviderAttestation.ProviderAsns[i].AddressFamilyIdentifier)
				return nil, errors.New("ProviderAsns AddressFamilyIdentifier is not 0x01 or 0x02")
			}
			addressFamily := null.IntFrom(afi.Int64())
			belogs.Debug("convertAsProviderAttestationToCustomerAsns(): ProviderAsns[i] addressFamily:", addressFamily)
			providerAsn.AddressFamily = addressFamily
		}
		providerAsn.ProviderAsn = uint64(asProviderAttestation.ProviderAsns[i].ProviderAsn)
		providerAsns = append(providerAsns, providerAsn)
	}
	customerAsn.ProviderAsns = providerAsns
	customerAsns = append(customerAsns, customerAsn)
	belogs.Debug("convertAsProviderAttestationToCustomerAsns(): customerAsns:", jsonutil.MarshalJson(customerAsns))

	return customerAsns, nil
}
