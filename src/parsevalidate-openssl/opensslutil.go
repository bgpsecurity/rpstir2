package openssl

import (
	"sort"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func convertAsProviderAttestationToCustomerAsns(asProviderAttestation AsProviderAttestation) (customerAsns []model.CustomerAsn, err error) {
	belogs.Debug("convertAsProviderAttestationToCustomerAsns(): asProviderAttestation:", jsonutil.MarshalJson(asProviderAttestation))

	customerAsns = make([]model.CustomerAsn, 0)
	customerAsn := model.CustomerAsn{}
	customerAsn.Version = uint64(asProviderAttestation.Version.Version)
	customerAsn.CustomerAsn = uint64(asProviderAttestation.CustomerAsn)
	providerAsns := make([]uint64, 0)
	for i := range asProviderAttestation.ProviderAsns {
		providerAsns = append(providerAsns, uint64(asProviderAttestation.ProviderAsns[i]))
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

	belogs.Info("convertAsProviderAttestationToCustomerAsns(): customerAsns:", jsonutil.MarshalJson(customerAsns))

	return customerAsns, nil
}
