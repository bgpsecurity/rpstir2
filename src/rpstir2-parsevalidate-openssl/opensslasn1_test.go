package openssl

import (
	"fmt"
	"testing"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseSigModelByOpensslResults(t *testing.T) {
	sigModel := model.SigModel{}
	results, err := fileutil.ReadFileToLines(`F:\share\我的坚果云\Go\go-study\src\asn1sig2\asn1parsechecksig.txt`)
	fmt.Println(jsonutil.MarshalJson(results), err)
	err = ParseSigModelByOpensslResults(results, &sigModel)
	fmt.Println(jsonutil.MarshalJson(sigModel), err)

}
