// Helper file for global commands
package game

import (
	"fmt"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

type CommandResult int

const (
	CommandNone CommandResult = iota
	CommandQuit
	CommandSceneMain
	CommandSceneGame
	CommandSceneScoreList
	CommandSceneLeaderboard
)

var exitCommands = map[CommandResult]bool{
	CommandQuit:             true,
	CommandSceneMain:        true,
	CommandSceneGame:        true,
	CommandSceneScoreList:   true,
	CommandSceneLeaderboard: true,
}

var commands = map[string]string{
	":q":           "quit",
	":help":        "show this help",
	":game":        "go to game scene",
	":main":        "go to main scene",
	":scores":      "go to ScoreList scene",
	":leaderboard": "go to leaderboard",
}

func HandleCommands(input string, shell *term.Terminal) (handled bool, result CommandResult) {
	if len(input) > 0 && input[0] == ':' {
		switch input {
		case ":q":
			shell.Write([]byte("Goodbye!\n"))
			return true, CommandQuit
		case ":help":
			var helpText string
			for cmd, desc := range commands {
				helpText += fmt.Sprintf("%s - %s\n", cmd, desc)
			}
			shell.Write([]byte("Commands:\n" + helpText))
			return true, CommandNone
		case ":main":
			return true, CommandSceneMain
		case ":game":
			return true, CommandSceneGame
		case ":scores":
			return true, CommandSceneScoreList
		case ":leaderboard":
			return true, CommandSceneLeaderboard
		default:
			shell.Write([]byte("Unknown command: " + input + "\n"))
			return true, CommandNone
		}
	}
	return false, CommandNone
}

func SafeReadInput(shell *term.Terminal, s glider.Session, p *Player) (string, Scene, bool) {
	for {
		input, err := shell.ReadLine()
		if err != nil {
			s.Close()
			return "", nil, true
		}

		handled, result := HandleCommands(input, shell)
		if handled {
			if exitCommands[result] {
				shell.Write([]byte("\033[2J\033[H")) // Clear screen, just to be sure
			}
			switch result {
			case CommandQuit:
				s.Close()
				return "", nil, true
			case CommandSceneMain:
				return "", sceneMain, true
			case CommandSceneGame:
				return "", sceneGame, true
			case CommandSceneScoreList:
				return "", sceneScoreList, true
			case CommandSceneLeaderboard:
				return "", sceneLeaderboard, true
			}
			continue // command handled, wait for next input
		}

		return input, nil, false
	}
}
