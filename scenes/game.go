package scenes

import (
	"fmt"
	"sort"
	"ssh-battle/player"
	"ssh-battle/util"
	"time"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func Game(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	green := "\033[32m"
	reset := "\033[0m"

	shell.Write([]byte("Welcome to the typing scene!\n"))
	shell.Write([]byte("Type :q to quit or :help to find more scenes.\n\n"))

	sentence := util.GetSentences()
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
	score := player.ScoreCalculation(sentence, input, elapsed)
	p.Scores = append(p.Scores, score)
	player.SaveScore(p.ID, score)

	last := p.Scores[len(p.Scores)-1]
	fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %d\nTP: %.2f\n\n",
		*last.Accuracy,
		*last.WPM,
		*last.Duration,
		*last.TP,
	)

	shell.Write([]byte("Press Enter to show score list...\n"))
	shell.ReadLine()

	return ScoreList
}

func ScoreList(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write([]byte("Welcome to the score list scene!\n"))
	shell.Write([]byte("Type :q to quit or :help to find more scenes.\n\n"))
	
	sort.Slice(p.Scores, func (i, j int) bool {
		return *p.Scores[i].TP > *p.Scores[j].TP
	})
	for i, score := range p.Scores {
		if i >= 5 {
			break
		}
		fmt.Fprint(s, "----------\n")
		fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %d\nTP: %.2f\n\n",
			*score.Accuracy,
			*score.WPM,
			*score.Duration,
			*score.TP,
		)
	}

	shell.Write([]byte("Press Enter to go to player scene or enter any command\n"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return Game
}
