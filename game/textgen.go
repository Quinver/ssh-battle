package game

import (
	"log"
	"math/rand"
	"strings"
	"time"
)

func getSentences(n int) []string {
	words, err := GetWordsFromDB()
	if err != nil {
		log.Fatal(err)
	}

	if len(words) == 0 {
		log.Fatal("no words available from DB")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano())) // local rand.Rand instance
	result := make([]string, 0, n)

	for range n {
		length := r.Intn(5) + 10 // sentence length 4-9 words
		sentenceWords := make([]string, length)
		for j := range length {
			sentenceWords[j] = words[r.Intn(len(words))]
		}
		sentence := strings.Join(sentenceWords, " ")
		result = append(result, sentence)
	}

	return result
}
