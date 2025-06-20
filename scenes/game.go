package scenes

import (
	"fmt"
	"sort"
	"ssh-battle/player"
	"ssh-battle/util"
	"time"

	glider "github.com/gliderlabs/ssh"
)

func Game(s glider.Session, p *player.Player) Scene {
	shell := p.Shell
	clearTerminal(shell)

	// Header
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 🎮 \033[1;38;5;51mSingle Player Typing Game\033[0m\033[38;5;45m                 │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to start typing\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :q to quit or :help for more commands\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mReady:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────\033[0m\n"))
	shell.Write([]byte("\033[38;5;46mPress Enter when you're ready...\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))

	sentence := util.GetSentences()
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	shell.Write([]byte("\033[38;5;252m\n" + sentence + "\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
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
	shell.Write([]byte("\033[38;5;229m\nYour Results:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────────\033[0m\n"))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mAccuracy: \033[38;5;51m%.2f%%\033[0m\n", *last.Accuracy))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mWPM: \033[38;5;51m%.1f\033[0m\n", *last.WPM))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mTime: \033[38;5;51m%d seconds\033[0m\n", *last.Duration))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mTP Score: \033[38;5;51m%.2f\033[0m\n\n", *last.TP))

	shell.Write([]byte("\033[38;5;46mPress Enter to view your score list...\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
	shell.ReadLine()

	return ScoreList
}

func ScoreList(s glider.Session, p *player.Player) Scene {
	shell := p.Shell
	clearTerminal(shell)

	// Header
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 📊 \033[1;38;5;51mYour Top Scores\033[0m\033[38;5;45m                           │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to return to game\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :q to quit or :help for more commands\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mTop 5 Scores:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))

	sort.Slice(p.Scores, func(i, j int) bool {
		return *p.Scores[i].TP > *p.Scores[j].TP
	})

	if len(p.Scores) == 0 {
		shell.Write([]byte("\033[38;5;248mNo scores yet. Play a game to start!\033[0m\n\n"))
	} else {
		for i, score := range p.Scores {
			if i >= 5 {
				break
			}
			rankColor := "\033[38;5;252m"
			if i == 0 {
				rankColor = "\033[38;5;226m" // Gold
			} else if i == 1 {
				rankColor = "\033[38;5;250m" // Silver
			} else if i == 2 {
				rankColor = "\033[38;5;172m" // Bronze
			}

			shell.Write(fmt.Appendf(nil, "%s#%d\033[38;5;252m ────────────────────\033[0m\n", rankColor, i+1))
			shell.Write(fmt.Appendf(nil, "\033[38;5;248mAccuracy: \033[38;5;51m%.2f%%\033[0m\n", *score.Accuracy))
			shell.Write(fmt.Appendf(nil, "\033[38;5;248mWPM: \033[38;5;51m%.1f\033[0m\n", *score.WPM))
			shell.Write(fmt.Appendf(nil, "\033[38;5;248mTime: \033[38;5;51m%d seconds\033[0m\n", *score.Duration))
			shell.Write(fmt.Appendf(nil, "\033[38;5;248mTP Score: \033[38;5;51m%.2f\033[0m\n\n", *score.TP))
		}
	}

	shell.Write([]byte("\033[38;5;46mPress Enter to return to game...\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return Game
}
