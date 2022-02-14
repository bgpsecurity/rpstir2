package openssl

import (
	"fmt"
	"os"
	"testing"
)

func TestParseAsnModelByOpensslResults(t *testing.T) {
	//pathName := `./`
	//fileName := `apnic-rpki-root-iana-origin.cer`
	fmt.Println(os.Args)
	filePath := os.Args[len(os.Args)-1]
	results, err := GetResultsByOpensslX509(filePath)
	fmt.Println(results, err)
	asnModel, err := ParseAsnModelByOpensslResults(results)
	fmt.Println(asnModel, err)
}
