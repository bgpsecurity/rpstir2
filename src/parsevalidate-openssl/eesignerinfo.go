package openssl

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
)

//Try to store the error in statemode instead of returning err
func ParseEeCertModel(eeCertFile string, fileByte []byte, eeCertStart int, eeCertEnd int) (eeCertModel model.EeCertModel, err error) {
	start := time.Now()
	belogs.Debug("ParseEeCertModel():  eeCertFile:", eeCertFile, "  len(fileByte):", len(fileByte),
		"   eeCertStart:", eeCertStart, "   eeCertEnd:", eeCertEnd)
	eeCertModel.EeCertStart = uint64(eeCertStart)
	eeCertModel.EeCertEnd = uint64(eeCertEnd)
	err = ParseEeCertModelByX509(fileByte, &eeCertModel)
	if err != nil {
		belogs.Error("ParseEeCertModel():ParseEeCertModelByX509 err:", err)
		return eeCertModel, err
	}

	results, err := opensslutil.GetResultsByOpensslX509(eeCertFile)
	if err != nil {
		belogs.Error("ParseEeCertModel(): GetResultsByOpensslX509 fail, eeCertFile:", eeCertFile, err)
		return eeCertModel, err
	}
	belogs.Debug("ParseEeCertModel(): GetResultsByOpensslX509 eeCertFile:", eeCertFile, "  len(results):", len(results))

	//keyusage ,critical
	eeCertModel.KeyUsageModel.Critical, eeCertModel.KeyUsageModel.KeyUsageValue, err = ParseKeyUsageModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseEeCertModel(): ParseKeyUsageModelByOpensslResults fail, eeCertFile:", eeCertFile, err)
		return eeCertModel, err
	}

	// AIA SIA
	_, eeCertModel.SiaModel, err = ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("ParseEeCertModel(): ParseAiaModelSiaModelByOpensslResults fail, eeCertFile:", eeCertFile, err)
		return eeCertModel, err
	}

	belogs.Debug("ParseEeCertModel(): eeCertModel:", jsonutil.MarshalJson(eeCertModel), "  time(s):", time.Since(start))
	return eeCertModel, nil
}
