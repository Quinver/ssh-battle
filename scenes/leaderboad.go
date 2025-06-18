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
	shell := term.NewTerminal(s, ">")
	clearTerminal(shell)

	shell.Write([]byte("Welcome to the leaderboard!\n"))
	shell.Write([]byte("Type :q to quit or :help to find more scenes.\n\n"))

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
			return Main
		}

		leaderboard = append(leaderboard, entry)
	}

	fmt.Fprint(s, " Rank | Player         | TP     | Accuracy | WPM   | Time\n")
	fmt.Fprint(s, "------------------------------------------------------------\n")
	for i, entry := range leaderboard {
		fmt.Fprintf(s, " %4d | %-14s | %6.2f | %8.2f%% | %5.1f | %4ds\n",
			i+1,
			entry.PlayerName,
			*entry.Score.TP,
			*entry.Score.Accuracy,
			*entry.Score.WPM,
			*entry.Score.Duration,
		)
	}
	fmt.Fprintln(s)

	if err := rows.Err(); err != nil {
		shell.Write([]byte("Error reading leaderboard rows\n"))
		return Main
	}

	shell.Write([]byte("Press Enter to go to game scene or enter any command\n"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return Game
}
