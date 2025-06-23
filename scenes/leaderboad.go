package scenes

import (
	"fmt"
	"log"
	"ssh-battle/data"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
)

func Leaderboard(s glider.Session, p *player.Player) Scene {
	shell := p.Shell
	clearTerminal(shell)

	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 📊 \033[1;38;5;51mLeaderboard - Top Players\033[0m\033[38;5;45m                 │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to return to game\n"))
	shell.Write([]byte("\033[38;5;248m• Type :q to quit or :help for more commands\033[0m\n\n"))

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
		shell.Write([]byte("Can't find info\n"))
		log.Print(err)
		return Game
	}
	defer rows.Close()

	var leaderboard []player.LeaderboardEntry

	for rows.Next() {
		var entry player.LeaderboardEntry
		if err := rows.Scan(&entry.PlayerName, &entry.Score.TP, &entry.Score.Accuracy, &entry.Score.WPM, &entry.Score.Duration); err != nil {
			return Game
		}
		leaderboard = append(leaderboard, entry)
	}

	if len(leaderboard) == 0 {
		shell.Write([]byte("\033[38;5;248mNo leaderboard data available yet.\033[0m\n\n"))
	} else {
		shell.Write([]byte("\033[38;5;45m┌────────┬─────────────┬──────────┬───────┬───────────┬───────────┐\033[0m\n"))
		shell.Write([]byte("\033[38;5;45m│ Rank   │ Player      │ Accuracy │ WPM   │ Time (s)  │ TP Score  │\033[0m\n"))
		shell.Write([]byte("\033[38;5;45m├────────┼─────────────┼──────────┼───────┼───────────┼───────────┤\033[0m\n"))

		for i, entry := range leaderboard {
			rankColor := "\033[38;5;252m" // default grey
			switch i {
			case 0:
				rankColor = "\033[38;5;226m" // gold
			case 1:
				rankColor = "\033[38;5;250m" // silver
			case 2:
				rankColor = "\033[38;5;172m" // bronze
			}

			playerName := entry.PlayerName
			if len(playerName) > 11 {
				playerName = playerName[:11]
			}

			row := fmt.Sprintf(
				"%s│ %-6d │ %-11s │ %8.2f │ %5.1f │ %9d │ %9.2f │\033[0m\n",
				rankColor,
				i+1,
				playerName,
				*entry.Score.Accuracy,
				*entry.Score.WPM,
				*entry.Score.Duration,
				*entry.Score.TP,
			)
			shell.Write([]byte(row))
		}

		shell.Write([]byte("\033[38;5;45m└────────┴─────────────┴──────────┴───────┴───────────┴───────────┘\033[0m\n\n"))
	}

	shell.Write([]byte("\033[38;5;46mPress Enter to return to game...\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return Game
}
