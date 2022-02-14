package sys

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
)

func exportRtrForManrs() (rtrForManrss []RtrForManrs, err error) {

	rtrForManrss, err = exportRtrForManrsDb()
	if err != nil {
		belogs.Error("exportRtrForManrs(): exportRtrForManrsDb, fail:", err)
		return nil, err
	}
	belogs.Info("exportRtrForManrs(): len(rtrForManrss):", len(rtrForManrss))

	for i := range rtrForManrss {
		address := rtrForManrss[i].Address
		prefixLength := rtrForManrss[i].PrefixLength
		addressFill, err := iputil.FillAddressWithZero(address, iputil.GetIpType(address))
		if err != nil {
			belogs.Error("exportRtrForManrs(): FillAddressWithZero, fail:", jsonutil.MarshalJson(rtrForManrss[i]), err)
			return nil, err
		}
		rtrForManrss[i].Prefix = addressFill + "/" + convert.ToString(prefixLength)
		belogs.Debug("exportRtrForManrs(): rtrForManrss[i].Prefix:", rtrForManrss[i].Prefix)
	}
	return rtrForManrss, nil
}
