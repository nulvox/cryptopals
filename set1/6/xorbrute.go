package main

import (
	"encoding/hex"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

type Score struct {
	Score  int
	Length int
}

func testKey(key string, text string) []byte {
	// key := os.Args[1]
	// text := strings.Join(os.Args[2:], " ")
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
	return output
}

func checkHamDist(string1 string, string2 string) int {
	minLen := len(string1)
	maxLen := len(string2)
	if len(string2) < minLen {
		minLen = len(string2)
		maxLen = len(string1)
	}
	bytes1 := []byte(string1)
	bytes2 := []byte(string2)
	dist := 0
	for i := 0; i < maxLen; i++ {
		a := byte(0)
		b := byte(0)
		if i < len(bytes1) {
			a = bytes1[i]
		}
		if i < len(bytes2) {
			b = bytes2[i]
		}
		c := a ^ b
		dist += bits.OnesCount8(c)
	}
	return dist
}

func scoreChunk(guess string) int {
	for _, c := range guess {
		if c == 0 || c > 127 {
			return 0
		}
	}

	bigrams := []string{"th", "he", "in", "er", "an", "re", "on", "at", "en", "nd", "ti", "es", "or", "te", "of", "ed", "is", "it", "al", "ar"}

	lower := strings.ToLower(guess)
	score := 0

	for _, bigram := range bigrams {
		score += strings.Count(lower, bigram) * 5
	}

	// Bonus for lowercase letters (more natural English)
	for _, c := range guess {
		if c >= 'a' && c <= 'z' {
			score++
		}
	}

	return score
}

func crack(chunk string) string {
	// this function needs to crack a string with
	// a 1-byte xor cypher
	highscore := 0
	output := ""
	finalKey := 0
	for key := 0; key <= 0xFF; key++ {
		plainCandidate := ""
		for i := 0; i < len(chunk); i++ {
			plainCandidate += string(int(chunk[i]) ^ key)
		}
		score := scoreChunk(plainCandidate)
		if score > highscore {
			highscore = score
			finalKey = key
		}
	}
	return finalKey
}

func main() {
	// string1 := "this is a test"
	// string2 := "wokka wokka!!!"
	fmt.Printf("%d", checkHamDist(string1, string2))

	maxKeyLen, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Print("First arg must be a decimal number")
		panic(err)
	}
	b64File := os.Args[2]
	cryptText := b64decode(readfile(b64File))
	bestScore := Score{Length: maxKeyLen, Score: maxKeyLen * 8}
	goodScore := Score{Length: maxKeyLen, Score: maxKeyLen * 8}
	for keysize := 1; keysize <= maxKeyLen; keysize++ {
		if keysize*4 > len(cryptText) {
			break
		}
		first := cryptText[:keysize]
		second := cryptText[keysize : keysize*2]
		sizeScore := checkHamDist(first, second) / keysize
		if sizeScore < goodScore.Score {
			goodScore = Score{Length: keysize, Score: sizeScore}
			third := cryptText[keysize*2 : keysize*3]
			forth := cryptText[keysize*3 : keysize*4]
			score2 := checkHamDist(third, forth) / keysize
			averageScore := sizeScore + score2/2
			if averageScore < bestScore.Score {
				bestScore = Score{Length: keysize, Score: averageScore}
			}
		}
	}
	var rawChunks []string
	for chunkStart := 0; chunkStart < len(cryptText); chunkStart += bestScore.Length {
		rawChunks = append(rawChunks, cryptText[chunkStart:chunkStart+bestScore.Length])
	}
	var adjustedChunks []string
	for i := 0; i < len(rawChunks); i++ {
		adjustedChunks[i] = ""
	}
	// shuffle the blocks so we have them sepparated by key chars
	for chunk := 0; chunk < len(rawChunks); chunk++ {
		for chunkChar := 0; chunkChar < bestScore.Length; chunkChar++ {
			adjustedChunks[chunk] += rawChunks[chunkChar][chunk]
		}
	}
	key := ""
	for chunk := 0; chunk < len(adjustedChunks); chunk++ {
		keyPart := crack(adjustedChunks[chunk])
		key += keyPart
	}
	// @TODO, actually use the key and print the answer nicely.
	fmt.Print(key)
}
