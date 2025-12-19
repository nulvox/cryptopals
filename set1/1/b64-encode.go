package main

import (
	"encoding/hex"
	"fmt"
)

const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func encode(data []byte) string {
	if len(data) < 1 {
		return ""
	}
	result := make([]byte, 0, ((len(data)+2)*4)/3)

	// process 3 byte chunks
	for i := 0; i < len(data); i += 3 {
		b1 := data[i]
		var b2, b3 byte
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		// turn 3 bytes into 4 b64 chars
		result = append(result, base64Chars[b1>>2])
		result = append(result, base64Chars[((b1&0x03)<<4)|(b2>>4)])
		result = append(result, base64Chars[((b2&0xF)<<2)|(b3>>6)])
		result = append(result, base64Chars[b3&0x3f])
	}
	switch len(data) % 3 {
	case 1:
		result[len(result)-2] = '='
		result[len(result)-1] = '='
	case 2:
		result[len(result)-1] = '='
	}
	return string(result)
}

func main() {
	hexStr := "49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d"
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(encode(bytes))
}
