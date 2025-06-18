package scenes

import (
	"fmt"
	"log"
	"ssh-battle/data"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func Leaderboard(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "")
	clearTerminal(shell)

	// Header
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 🏆 \033[1;38;5;51mSSH Battle Leaderboard\033[0m\033[38;5;45m                     │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to go to game scene\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :q to quit or :help for more commands\033[0m\n\n"))

	// Leaderboard title
	shell.Write([]byte("\033[38;5;229mTop 10 Players:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m───────────────\033[0m\n"))

	rows, err := data.DB.Query(`
		SELECT 
			 p.username,
			 s.tp,
			 s.accuracy,
			 s.wpm,
			 s.duration
		FROM players p
		JOIN scores s ON p.id = s.player_id
		ORDER BY s.tp DESC
		LIMIT 10;
	`)
	if err != nil {
		shell.Write([]byte("\033[38;5;196mError: Cannot fetch leaderboard data\033[0m\n"))
		log.Print(err)
		return Game
	}
	defer rows.Close()

	var leaderboard []player.LeaderboardEntry

	for rows.Next() {
		var entry player.LeaderboardEntry
		if err := rows.Scan(&entry.PlayerName, &entry.Score.TP, &entry.Score.Accuracy, &entry.Score.WPM, &entry.Score.Duration); err != nil {
			shell.Write([]byte("\033[38;5;196mError: Cannot read leaderboard entry\033[0m\n"))
			return Main
		}

		leaderboard = append(leaderboard, entry)
	}

	// Table header
	shell.Write([]byte("\033[1;38;5;51m Rank │ Player         │ TP     │ Accuracy │ WPM   │ Time  \033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────┼────────────────┼────────┼──────────┼───────┼───────\033[0m\n"))

	// Table rows
	for i, entry := range leaderboard {
		rankColor := "\033[38;5;252m"
		if i == 0 {
			rankColor = "\033[38;5;226m" // Gold for 1st
		} else if i == 1 {
			rankColor = "\033[38;5;250m" // Silver for 2nd
		} else if i == 2 {
			rankColor = "\033[38;5;172m" // Bronze for 3rd
		}

		shell.Write(fmt.Appendf(nil, "%s%4d \033[38;5;252m│ \033[38;5;248m%-14s │ %6.2f │ %7.2f%% │ %5.1f │ %4ds\033[0m\n",
			rankColor, i+1, entry.PlayerName, *entry.Score.TP, *entry.Score.Accuracy, *entry.Score.WPM, *entry.Score.Duration))
	}

	if err := rows.Err(); err != nil {
		shell.Write([]byte("\033[38;5;196mError reading leaderboard rows\033[0m\n"))
		return Main
	}

	shell.Write([]byte("\n\033[38;5;208m> \033[0m"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return Game
}
