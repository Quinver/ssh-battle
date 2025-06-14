package game

import (
	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"

	"fmt"
	"time"
)

// Scene function type for switching between scenes
type Scene func(glider.Session, *Player) Scene

func sceneMain(s glider.Session, p *Player) Scene {
	shell := term.NewTerminal(s, "> ")
	shell.Write([]byte("You're in the main scene. Type :help to find more scenes\n"))

	for {
		_, nextScene, done := SafeReadInput(shell, s, p)
		if done {
			return nextScene
		}
	}
}

func sceneGame(s glider.Session, p *Player) Scene {
	shell := term.NewTerminal(s, "> ")

	green := "\033[32m"
	reset := "\033[0m"

	shell.Write([]byte("Welcome to the typing scene!\n"))
	shell.Write([]byte("Type :q to quit or :help to find more scenes.\n\n"))

	for _, sentence := range getSentences(1) {
		shell.Write([]byte("Press Enter when you're ready...\n"))
		_, nexScene, done := SafeReadInput(shell, s, p)
		if done {
			return nexScene
		}

		shell.Write([]byte(green + sentence + reset + "\n"))
		start := time.Now()

		input, nextScene, done := SafeReadInput(shell, s, p)
		if done {
			return nextScene
		}

		elapsed := time.Since(start)
		score := scoreCalculation(sentence, input, elapsed)
		p.Scores = append(p.Scores, score)

		last := p.Scores[len(p.Scores)-1]
		fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %.2fs\n\n",
			*last.Accuracy,
			*last.WPM,
			last.Time.Seconds())
	}

	return nil
}
