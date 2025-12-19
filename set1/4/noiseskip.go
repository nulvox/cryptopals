package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
)

type Entry struct {
	Line  int
	Key   string
	Crypt string
	Clear string
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func decodeGuess(key int, bytes []byte) string {
	result := make([]byte, len(bytes))
	for i := 0; i < len(bytes); i++ {
		result[i] = bytes[i] ^ byte(key)
	}
	return string(result)
}

func rateGuess(guess string) int {
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

func main() {
	var results []Entry
	highscore := 0
	lines, err := readLines("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	for cryptindex := 0; cryptindex < len(lines); cryptindex++ {
		bytes, err := hex.DecodeString(lines[cryptindex])
		if err != nil {
			panic(err)
		}
		for key := 0; key < 0xFF; key++ {
			guess := decodeGuess(key, bytes)
			score := rateGuess(guess)
			if score == highscore {
				results = append(results, Entry{Line: cryptindex, Key: string(key), Crypt: lines[cryptindex], Clear: guess})
			} else if score > highscore {
				results = nil
				highscore = score
				results = append(results, Entry{Line: cryptindex, Key: string(key), Crypt: lines[cryptindex], Clear: guess})
			}
		}
	}
	for i := 0; i < len(results); i++ {
		fmt.Printf("With the key %s on line %d we get\n %s \nfrom\n %s\n\n", results[i].Key, results[i].Line, results[i].Clear, results[i].Crypt)
	}
}
