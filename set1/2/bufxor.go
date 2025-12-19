package main

import (
	"encoding/hex"
	"fmt"
)

func xorbytes(a, b []byte) []byte {
	result := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func main() {
	input1 := "1c0111001f010100061a024b53535009181c"
	input2 := "686974207468652062756c6c277320657965"

	if len(input1) != len(input2) {
		fmt.Println("Error, input buffers are not of equal len")
	}
	bytes1, err1 := hex.DecodeString(input1)
	if err1 != nil {
		panic(err1)
	}
	bytes2, err2 := hex.DecodeString(input2)
	if err2 != nil {
		panic(err2)
	}
	fmt.Printf("%X\n", xorbytes(bytes1, bytes2))
}
