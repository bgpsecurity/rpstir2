package openssl

import (
	"fmt"
	"testing"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/opensslutil"
	"github.com/cpusoft/goutil/osutil"
)

func TestParseSigModelByOpensslResults(t *testing.T) {
	sigModel := model.SigModel{}
	results, err := fileutil.ReadFileToLines(`asn1parsechecksig.txt`)
	fmt.Println(jsonutil.MarshalJson(results), err)
	err = ParseSigModelByOpensslResults(results, &sigModel)
	fmt.Println(jsonutil.MarshalJson(sigModel), err)

}

func TestParseAsaModelByOpensslResults(t *testing.T) {
	m := make(map[string]string, 0)
	m[".asa"] = ".asa"
	fs, err := osutil.GetFilesInDir(`G:\Download\cert\aspa\`, m)
	fmt.Println(fs, err)
	for _, file := range fs {
		fmt.Println("---------file:", file)
		asaModel := &model.AsaModel{}
		results, err := opensslutil.GetResultsByOpensslAns1(`G:\Download\cert\aspa\` + file)
		fmt.Println("len(results):", len(results), err)
		err = ParseAsaModelByOpensslResults(results, asaModel)
		fmt.Println("asaModel:", jsonutil.MarshalJson(asaModel), err)
	}
}
