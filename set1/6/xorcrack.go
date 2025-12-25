package main

import (
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

type Solution struct {
	Score int
	Key   []byte
	Plain []byte
}

func readfile(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			panic("file not found")
		}
		panic(err)
	}

	if !info.Mode().IsRegular() {
		panic("not a plain file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func b64decode(s string) []byte {
	const b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	// Build reverse lookup table
	lookup := make(map[byte]int)
	for i := 0; i < len(b64); i++ {
		lookup[b64[i]] = i
	}

	var out []byte
	var buf, bits int

	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			break
		}

		val := lookup[s[i]]
		buf = (buf << 6) | val
		bits += 6

		if bits >= 8 {
			bits -= 8
			out = append(out, byte(buf>>bits))
			buf &= (1 << bits) - 1
		}
	}

	return out
}

func blockXor(key []byte, text []byte) []byte {
	var output []byte
	for i := 0; i < len(text); i++ {
		output = append(output, text[i]^key[i%len(key)])
	}
	return output
}

func checkHamDist(bytes1 []byte, bytes2 []byte) int {
	maxLen := len(bytes1)
	if len(bytes2) > maxLen {
		maxLen = len(bytes2)
	}
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

func findKeysizes(maxKeysize int, crypText []byte) []int {
	type keysizeScore struct {
		size  int
		score float64
	}

	var scores []keysizeScore

	for keysize := 2; keysize <= maxKeysize; keysize++ {
		if len(crypText) < keysize*4 {
			continue
		}

		// Compare multiple block pairs for better accuracy
		numBlocks := len(crypText) / keysize
		if numBlocks > 8 {
			numBlocks = 8 // Use up to 8 blocks
		}

		totalDist := 0
		comparisons := 0

		for i := 0; i < numBlocks-1; i++ {
			start1 := keysize * i
			end1 := start1 + keysize
			start2 := end1
			end2 := start2 + keysize

			if end2 > len(crypText) {
				break
			}

			block1 := crypText[start1:end1]
			block2 := crypText[start2:end2]
			totalDist += checkHamDist(block1, block2)
			comparisons++
		}

		if comparisons == 0 {
			continue
		}

		// Normalize by keysize and number of comparisons
		normalized := float64(totalDist) / float64(comparisons*keysize)
		scores = append(scores, keysizeScore{keysize, normalized})
	}

	// Sort by score (lower is better) and return top 3 candidates
	if len(scores) == 0 {
		return nil
	}

	// Simple bubble sort to find minimum
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score < scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Return top 3 keysizes
	var result []int
	limit := 3
	if len(scores) < limit {
		limit = len(scores)
	}
	for i := 0; i < limit; i++ {
		result = append(result, scores[i].size)
	}

	return result
}

func breakChunks(keysize int, crypText []byte) [][]byte {
	var rawChunks [][]byte
	for chunkStart := 0; chunkStart < len(crypText); chunkStart += keysize {
		end := chunkStart + keysize
		if end > len(crypText) {
			end = len(crypText)
		}
		rawChunks = append(rawChunks, crypText[chunkStart:end])
	}
	return rawChunks
}

func arrangeChunks(chunks [][]byte) [][]byte {
	if len(chunks) == 0 {
		return nil
	}

	keysize := len(chunks[0])
	adjustedChunks := make([][]byte, keysize)

	for chunkChar := 0; chunkChar < keysize; chunkChar++ {
		for chunk := 0; chunk < len(chunks); chunk++ {
			if chunkChar < len(chunks[chunk]) {
				adjustedChunks[chunkChar] = append(adjustedChunks[chunkChar], chunks[chunk][chunkChar])
			}
		}
	}
	return adjustedChunks
}

func crackKeyByte(crypChunk []byte) byte {
	bestKey := byte(1)
	highscore := -999999

	for key := 1; key <= 0xFF; key++ {
		keySlice := []byte{byte(key)}
		plainChunk := blockXor(keySlice, crypChunk)

		score := 0
		nonPrintable := 0

		for _, c := range plainChunk {
			if c > 127 {
				score = -1000
				break
			}
			// Allow common whitespace
			if c < 32 && c != '\n' && c != '\t' && c != '\r' {
				nonPrintable++
			}
			// Score printable chars
			if c >= 32 && c <= 126 {
				score++
			}
			// Bonus for common letters
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == ' ' {
				score += 2
			}
			// Extra bonus for very common letters
			if c == 'e' || c == 't' || c == 'a' || c == 'o' || c == 'i' || c == 'n' || c == ' ' {
				score++
			}
		}

		// Penalize excessive non-printable chars
		if float64(nonPrintable)/float64(len(plainChunk)) > 0.1 {
			score -= 500
		}

		if score > highscore {
			highscore = score
			bestKey = byte(key)
		}
	}

	return bestKey
}

func crackKey(keysizes []int, crypText []byte) []Solution {
	var solutions []Solution
	var emptyBytes []byte
	for keysizeOffset := 0; keysizeOffset < len(keysizes); keysizeOffset++ {
		rawChunks := breakChunks(keysizes[keysizeOffset], crypText)
		adjustedChunks := arrangeChunks(rawChunks)
		var key []byte
		for chunkOffset := 0; chunkOffset < len(adjustedChunks); chunkOffset++ {
			key = append(key, crackKeyByte(adjustedChunks[chunkOffset]))
		}
		solutions = append(solutions, Solution{Key: key, Plain: emptyBytes, Score: 0})
	}
	return solutions
}

func decryptSolution(sol Solution, crytpText []byte) Solution {
	sol.Plain = blockXor(sol.Key, crytpText)
	return sol
}

func scoreGuess(sol Solution) int {
	// Count non-printable chars (excluding common whitespace)
	nonPrintable := 0
	for _, c := range sol.Plain {
		if c > 127 {
			return -1000 // Definitely invalid
		}
		// Allow newline, tab, carriage return
		if c < 32 && c != '\n' && c != '\t' && c != '\r' {
			nonPrintable++
		}
	}

	// If more than 5% non-printable, probably garbage
	if float64(nonPrintable)/float64(len(sol.Plain)) > 0.05 {
		return -1000
	}

	bigrams := []string{"th", "he", "in", "er", "an", "re", "on", "at", "en", "nd", "ti", "es", "or", "te", "of", "ed", "is", "it", "al", "ar"}

	lower := strings.ToLower(string(sol.Plain))
	score := 0

	for _, bigram := range bigrams {
		score += strings.Count(lower, bigram) * 5
	}

	// Bonus for lowercase letters (more natural English)
	for _, c := range sol.Plain {
		if c >= 'a' && c <= 'z' {
			score++
		}
	}

	return score
}

func solveAnswers(answers []Solution, crypText []byte) []Solution {
	highscore := -1
	var output []Solution
	for solution := 0; solution < len(answers); solution++ {
		decryptedGuess := decryptSolution(answers[solution], crypText)
		score := scoreGuess(decryptedGuess)

		if score < 0 {
			continue
		}

		if score > highscore {
			output = nil
			highscore = score
		}
		if score == highscore {
			decryptedGuess.Score = score
			output = append(output, decryptedGuess)
		}
	}

	fmt.Printf("Final highscore: %d, num solutions: %d\n", highscore, len(output))
	return output
}

func reportAnswers(answers []Solution) {
	for answer := 0; answer < len(answers); answer++ {
		fmt.Printf("Key: %s\n", string(answers[answer].Key))
		fmt.Printf("Score: %d\n", answers[answer].Score)
		fmt.Println("Message:")
		fmt.Println("---")
		fmt.Println(string(answers[answer].Plain))
		fmt.Println("---")
		fmt.Println()
	}
}

func main() {
	maxKeysize, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Print("First arg must be a decimal number")
		panic(err)
	}
	b64File := os.Args[2]
	fileContent := readfile(b64File)

	// Debug: check file content
	fmt.Printf("File length: %d\n", len(fileContent))
	fmt.Printf("First 100 chars: %q\n", fileContent[:100])

	// Remove newlines from base64 input
	fileContent = strings.ReplaceAll(fileContent, "\n", "")
	fileContent = strings.ReplaceAll(fileContent, "\r", "")

	fmt.Printf("After cleanup length: %d\n", len(fileContent))

	crypText := b64decode(fileContent)
	fmt.Printf("Decoded length: %d\n", len(crypText))
	fmt.Printf("First 20 bytes: %x\n", crypText[:20])

	keysizes := findKeysizes(maxKeysize, crypText)
	answers := crackKey(keysizes, crypText)
	answers = solveAnswers(answers, crypText)
	reportAnswers(answers)
}
