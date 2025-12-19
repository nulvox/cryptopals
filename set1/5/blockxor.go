package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func main() {
	key := os.Args[1]
	text := strings.Join(os.Args[2:], " ")
	textbytes, err := hex.DecodeString(text)
	if err != nil {
		textbytes = []byte(text)
	}
	keybytes := []byte(key)
	output := make([]byte, len(text))
	for i := 0; i < len(textbytes); i++ {
		keyoff := i % len(keybytes)
		output[i] = textbytes[i] ^ keybytes[keyoff]
	}
	fmt.Printf("%x", output)
}
