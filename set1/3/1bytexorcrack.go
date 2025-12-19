package main

import (
	"encoding/hex"
	"fmt"
)

func decodeGuess(key int, bytes []byte) []byte {
	result := make([]byte, len(bytes))
	for i := 0; i < len(bytes); i++ {
		result[i] = bytes[i] ^ byte(key)
	}
	return result
}

func rateGuesses(guesses [][]byte) (string, string) {
	// store the string versions of the messages
	outputs := make([]string, 0xFF)
	// count the number of alphabetic chars as a base score
	scores := make([]int, 0xFF)
	// keep track of the highscore
	highscore := 0
	a := byte('a')
	z := byte('z')
	for i := 0; i < 0xFF; i++ {
		// count letters
		count := 0
		for _, b := range guesses[i] {
			if b >= 0x41 && b <= 0x7A {
				count++
			}
			if b >= a && b <= z {
				count++
			}
		}
		// record highscores
		if count > highscore {
			highscore = count
		}
		// record all scores
		scores[i] = count
		// record all outputs
		outputs[i] = string(guesses[i])
	}
	// if multiple strings tie for highscore
	//  break the tie based on which word makes the most sense...
	//  somehow
	guess := ""
	key := 0
	for i := 0; i < 0xFF; i++ {
		if scores[i] == highscore {
			fmt.Println(outputs[i])
			guess = outputs[i]
			key = i
		}
	}
	keystr := fmt.Sprintf("%#X", key)
	return keystr, guess
}

func main() {
	hexinput := "1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736"
	bytes, err := hex.DecodeString(hexinput)
	if err != nil {
		panic(err)
	}
	// 16 guesses for each guess
	guesses := make([][]byte, 0xFF)
	for i := 0; i < 0xFF; i++ {
		guesses[i] = make([]byte, len(bytes))
		// Optional: fill with the value i
		for j := range guesses[i] {
			guesses[i][j] = 0
		}
	}
	for i := 0; i < 0xFF; i++ {
		key := i
		guesses[i] = decodeGuess(key, bytes)
	}
	bestkey, output := rateGuesses(guesses)
	fmt.Printf("The key is %s and the input was %s", bestkey, output)
}
