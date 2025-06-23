package player

import (
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"ssh-battle/data"
)

type Score struct {
	ID        *int
	PlayerID  *int
	Accuracy  *float64
	WPM       *float64
	TP        *float64
	Duration  *int
	CreatedAt *time.Time
}

type LeaderboardEntry struct {
	PlayerName string
	Score      Score
}

func ScoreCalculation(ref, pred string, elapsed time.Duration) Score {
	refChars := []rune(ref)
	predChars := []rune(pred)

	// totalRefChars := len(refChars)
	totalPredChars := len(predChars)

	minLen := min(len(predChars), len(refChars))

	correct := 0

	for i := range minLen {
		if refChars[i] == predChars[i] {
			correct++
		}
	}

	acc := AccuracyPerWord(ref, pred)

	secs := elapsed.Seconds()
	if secs == 0 {
		secs = 1 // Avoid division by zero
	}

	wpm := (60.0 * float64(totalPredChars/5.0)) / secs
	d := int(secs)

	tp := CalculateTP(acc, wpm, d)
	return Score{
		Accuracy: &acc,
		WPM:      &wpm,
		Duration: &d,
		TP:       &tp,
	}
}

func AccuracyPerWord(ref, pred string) float64 {
	refWords := strings.Fields(ref)
	predWords := strings.Fields(pred)

	totalChars := 0
	correctChars := 0

	minWords := min(len(refWords), len(predWords))

	for i := range minWords {
		refChars := []rune(refWords[i])
		predChars := []rune(predWords[i])
		minLen := min(len(refChars), len(predChars))

		for j := range minLen {
			if refChars[j] == predChars[j] {
				correctChars++
			}
		}
		totalChars += len(refChars)
	}

	// Account for extra words in ref (penalty for missing words)
	for i := minWords; i < len(refWords); i++ {
		totalChars += len([]rune(refWords[i]))
	}

	if totalChars == 0 {
		return 0
	}

	return (float64(correctChars) / float64(totalChars)) * 100
}

func CalculateTP(accuracy float64, wpm float64, duration int) float64 {
	const accWeight = 1.2
	const wpmWeight = 1.5
	const timeWeight = 0.8

	// Nerf extremely short and long sessions
	durFactor := math.Log10(float64(duration) + 10)

	tp := (math.Pow(accuracy, accWeight) * math.Pow(wpm, wpmWeight)) / (durFactor * 1000)

	return tp
}

var scoreMu sync.Mutex

func SaveScore(playerID int, score Score) error {
	scoreMu.Lock()
	defer scoreMu.Unlock()

	createdAt := time.Now()
	if score.CreatedAt != nil && !score.CreatedAt.IsZero() {
		createdAt = *score.CreatedAt
	}

	tx, err := data.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        INSERT INTO scores (player_id, accuracy, wpm, tp, duration, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `, playerID, score.Accuracy, score.WPM, score.TP, score.Duration, createdAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Player with id %d submitted a score", playerID)
	return nil
}
