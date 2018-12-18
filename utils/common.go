package utils

import (
	"encoding/hex"
	"fmt"
)

//16进制的字符表示和二进制字节表示互转
var (
	Bin2Hex = hex.EncodeToString
	Hex2Bin = hex.DecodeString
)

func CheckError(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
