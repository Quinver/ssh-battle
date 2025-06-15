package util

import (
	"log"
	"math/rand"
	"strings"
	"time"
	"ssh-battle/data"
)

func GetSentences(n int) []string {
	words, err := getWordsFromDB()
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


func getWordsFromDB() ([]string, error) {
	rows, err := data.DB.Query("SELECT word FROM words")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var DBWords []string
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			return nil, err
		}
		DBWords = append(DBWords, w)
	}
	return DBWords, nil
}
