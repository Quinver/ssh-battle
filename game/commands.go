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
)

var exitCommands = map[CommandResult]bool{
	CommandQuit:      true,
	CommandSceneMain: true,
	CommandSceneGame: true,
}

var commands = map[string]string{
	":q":    "quit",
	":help": "show this help",
	":game": "go to game scene",
	":main": "go to main scene",
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
		default:
			shell.Write([]byte("Unknown command: " + input + "\n"))
			return true, CommandNone
		}
	}
	return false, CommandNone
}

func ReadInput(shell *term.Terminal) (string, CommandResult) {
	for {
		input, err := shell.ReadLine()
		if err != nil {
			return "", CommandQuit // quit on error
		}

		handled, result := HandleCommands(input, shell)
		if handled {
			if exitCommands[result] {
				return "", result
			}
			continue // command handled but no quit, ask for input again
		}

		return input, CommandNone
	}
}

// Returns Input of user, if bool is true exit current loop
func SafeReadInput(shell *term.Terminal, s glider.Session, p *Player) (string, Scene, bool) {
	input, result := ReadInput(shell)

	switch result {
	case CommandQuit:
		s.Close()
		return "", nil, true
	case CommandSceneMain:
		return "", sceneMain, true
	case CommandSceneGame:
		return "", sceneGame, true
	default:
		return input, nil, false
	}
}
