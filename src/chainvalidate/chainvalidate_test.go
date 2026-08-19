package chainvalidate

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	_ "gonum.org/v1/gonum/graph/simple"
)

//func normalizeHexString(s string) string {
//	// 1. 转小写
//	s = strings.ToLower(s)
//	// 2. 移除所有空格
//	s = strings.ReplaceAll(s, " ", "")
//	// 3. 移除可能的 "0x" 前缀
//	s = strings.TrimPrefix(s, "0x")
//	// 4. 移除可能的 ":" 分隔符
//	s = strings.ReplaceAll(s, ":", "")
//	return s
//}

// HexToBinary 将十六进制字符串转换为二进制字符串
func HexToBinary(hexStr string) (string, error) {
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}

	// 将十六进制字符串转换为字节数组
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("hex string decode failed: %v", err)
	}

	// 将每个字节转换为 8 位二进制字符串，并拼接起来
	var binaryStrings []string
	for _, b := range bytes {
		binaryStrings = append(binaryStrings, fmt.Sprintf("%08b", b))
	}
	binaryStr := strings.Join(binaryStrings, "")

	return binaryStr, nil
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Test_chainValidate2(t *testing.T) {

	chainValidate(0)
}

// 本地捞数据
func Test_chainValidate3(t *testing.T) {

	chainValidate(0)
}
