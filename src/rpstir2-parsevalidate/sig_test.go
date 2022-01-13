package parsevalidate

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseValidateSig(t *testing.T) {

	file := `../../../go-study/src/asn1sig2/checklist.sig`
	sigModel, stateModel, err := ParseValidateSig(file)
	fmt.Println(jsonutil.MarshalJson(sigModel), jsonutil.MarshalJson(stateModel), err)

}
