package game

import (
	"fmt"
	"time"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)


func SessionStart(s glider.Session) {
	player := newPlayer(s)
	fmt.Fprintf(s, "%s Joined \n", player.Name)

	// keep session open
	for {
		inGame(s, player)
	}
}

func inGame(s glider.Session, p *Player) {
	shell := term.NewTerminal(s, "> ")

	// Text Colors
	green := "\033[32m"
	reset := "\033[0m"

	// Loop through sentences, goes to next sentence when user inputs new line
	for _, sentence := range getSentences(1) {
		shell.Write([]byte("Press Enter when you're ready...\n"))
		_, quit := ReadInput(shell)
		if quit {
			s.Close()
			return
		}

		start := time.Now()

		shell.Write([]byte(green + sentence + reset + "\n"))

		input, quit := ReadInput(shell)
		if quit {
			s.Close()
			return
		}

		elapsed := time.Since(start)

		score := scoreCalculation(sentence, input, elapsed)
		p.Scores = append(p.Scores, score)

		last := p.Scores[len(p.Scores)-1]
		fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %.2fs\n",
			*last.Accuracy,
			*last.WPM,
			last.Time.Seconds())

	}
}


