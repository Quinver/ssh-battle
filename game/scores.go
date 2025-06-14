package game

import (
	"strings"
	"time"
)

type Score struct {
	Accuracy *float64
	WPM      *float64
	Time     *time.Duration
}


func scoreCalculation(ref, pred string, elapsed time.Duration) Score {
	refWords := strings.Fields(ref)   // Rererence
	predWords := strings.Fields(pred) // Predicted

	totalRefWords := len(refWords)
	totalPredWords := len(predWords)
	correct := 0

	// Correct++ when the words in the same position of slice is the same
	for i := 0; i < totalRefWords && i < totalPredWords; i++ {
		if refWords[i] == predWords[i] {
			correct++
		}
	}

	acc := float64(correct) / float64(totalRefWords) * 100
	wpm := (60.0 * float64(totalPredWords)) / float64(elapsed.Seconds())

	return Score{
		Accuracy: &acc,
		WPM:      &wpm,
		Time:     &elapsed,
	}
}
