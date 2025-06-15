package game

import (
	"log"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"

	"fmt"
	"time"
)

// Scene function type for switching between scenes
type Scene func(glider.Session, *Player) Scene

func sceneMain(s glider.Session, p *Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write([]byte("You're in the main scene. Type :help to find more scenes\n"))

	for {
		_, nextScene, done := SafeReadInput(shell, s, p)
		if done {
			return nextScene
		}
	}
}

func sceneMultiplayerRoom(s glider.Session, p *Player, roomID string) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write(fmt.Appendf(nil, "Welcome to the room: %s\n\n", roomID))

	room := GetRoom(roomID)
	room.Join <- p
	defer func() {
		room.Leave <- p
	}()

	done := make(chan struct{})
	nextSceneCh := make(chan Scene, 1)

	go func() {
		for {
			select {
			case msg, ok := <-p.Messages:
				if !ok {
					return
				}
				shell.Write([]byte(msg + "\n"))
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			line, nextScene, finished := SafeReadInput(shell, s, p)
			if finished {
				close(done)
				nextSceneCh <- nextScene
				return
			}
			room.Broadcast <- RoomMessage{
				Sender:  p.Name,
				Content: fmt.Sprintf("[%s] %s", p.Name, line),
			}
		}
	}()

	for {
		select {
		case <-s.Context().Done():
			close(done)
			return nil
		case <-done:
			select {
			case nextScene := <-nextSceneCh:
				return nextScene
			default:
				return nil
			}
		}
	}
}

func sceneMultiplayerLobby(s glider.Session, p *Player) Scene {
	return sceneMultiplayerRoom(s, p, "main-lobby")
}

func sceneGame(s glider.Session, p *Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

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
		saveScore(p.ID, score)

		last := p.Scores[len(p.Scores)-1]
		fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %d\nTP: %.2f\n\n",
			*last.Accuracy,
			*last.WPM,
			*last.Duration,
			*last.TP,
		)

		shell.Write([]byte("Press Enter to show score list...\n"))
		shell.ReadLine()

	}

	return sceneScoreList
}

func sceneScoreList(s glider.Session, p *Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write([]byte("Welcome to the score list scene!\n"))
	shell.Write([]byte("Type :q to quit or :help to find more scenes.\n\n"))

	for _, score := range p.Scores {
		fmt.Fprint(s, "----------\n")
		fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %d\nTP: %.2f\n\n",
			*score.Accuracy,
			*score.WPM,
			*score.Duration,
			*score.TP,
		)
	}

	shell.Write([]byte("Press Enter to go to game scene or enter any command\n"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return sceneGame
}

func sceneLeaderboard(s glider.Session, p *Player) Scene {
	shell := term.NewTerminal(s, ">")
	clearTerminal(shell)

	shell.Write([]byte("Welcome to the leaderboard!\n"))
	shell.Write([]byte("Type :q to quit or :help to find more scenes.\n\n"))

	rows, err := db.Query(`
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
	}
	defer rows.Close()

	var leaderboard []LeaderboardEntry

	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.PlayerName, &entry.Score.TP, &entry.Score.Accuracy, &entry.Score.WPM, &entry.Score.Duration); err != nil {
			return sceneMain
		}

		leaderboard = append(leaderboard, entry)
	}

	for i, entry := range leaderboard {

		fmt.Fprint(s, "----------\n")
		fmt.Fprintf(s, "%d. %s\nAccuracy: %.2f%%\nWPM: %.1f\nTime: %d\nTP: %.2f\n\n",
			i+1,
			entry.PlayerName,
			*entry.Score.Accuracy,
			*entry.Score.WPM,
			*entry.Score.Duration,
			*entry.Score.TP,
		)
	}

	if err := rows.Err(); err != nil {
		shell.Write([]byte("Error reading leaderboard rows\n"))
		return sceneMain
	}

	shell.Write([]byte("Press Enter to go to game scene or enter any command\n"))
	_, nextScene, done := SafeReadInput(shell, s, p)
	if done {
		return nextScene
	}

	return sceneGame
}

func clearTerminal(shell *term.Terminal) {
	shell.Write([]byte("\033[2J\033[H")) // ANSI escape to clear screen and move cursor home
}
