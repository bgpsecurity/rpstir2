package model

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestRtrPdu(t *testing.T) {
	rtrSN := NewRtrSerialNotify()
	rtrSN.Length = 111
	rtrSN.SerialNumber = 123
	jsonutil.MarshalJson(rtrSN)
	fmt.Println(rtrSN)
	fmt.Println(rtrSN.Bytes())
	fmt.Println(hex.Dump(rtrSN.Bytes()))
	fmt.Println(convert.PrintBytes(rtrSN.Bytes(), 8))

}
