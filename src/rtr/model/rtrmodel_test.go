package model

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestRtrPdu(t *testing.T) {
	erm := NewRtrErrorReportModel(0, 1, nil, nil)
	fmt.Println(jsonutil.MarshalJson(erm))
	fmt.Println(hex.Dump(erm.Bytes()))
	fmt.Println(convert.PrintBytes(erm.Bytes(), 8))

}
