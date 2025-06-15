// Helper file for global commands
package scenes

import (
	"fmt"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

type Command struct {
	Description string
	Handler     func(shell *term.Terminal)
	NextScene   Scene // nil if no scene transition
	Quit        bool
}

var commandRegistry map[string]Command

func init() {
	commandRegistry = map[string]Command{
		":q": {
			Description: "quit",
			Handler: func(shell *term.Terminal) {
				shell.Write([]byte("Goodbye!\n"))
			},
			NextScene: nil,
			Quit: true,
		},
		":help": {
			Description: "show this help",
			Handler: func(shell *term.Terminal) {
				shell.Write([]byte("Commands:\n"))
				for cmd, data := range commandRegistry {
					shell.Write(fmt.Appendf(nil, "%s - %s\n", cmd, data.Description))
				}
			},
			NextScene: nil,
		},
		":main": {
			Description: "go to main scene",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Main,
		},
		":game": {
			Description: "go to game scene",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Game,
		},
		":lobby": {
			Description: "go to multiplayer lobby",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   MultiplayerLobby,
		},
		":scores": {
			Description: "go to ScoreList scene",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   ScoreList,
		},
		":leaderboard": {
			Description: "go to leaderboard",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Leaderboard,
		},
	}

	// Aliases
	commandRegistry[":exit"] = commandRegistry[":q"]
	commandRegistry[":quit"] = commandRegistry[":q"]
}

func SafeReadInput(shell *term.Terminal, s glider.Session, p *player.Player) (string, Scene, bool) {
	for {
		input, err := shell.ReadLine()
		if err != nil {
			s.Close()
			return "", nil, true
		}

		if cmd, ok := commandRegistry[input]; ok {
			cmd.Handler(shell)
			if cmd.Quit{
				s.Close()
				return "", nil, true
			}
			if cmd.NextScene != nil {
				shell.Write([]byte("\033[2J\033[H"))
				return "", cmd.NextScene, true
			}
			continue
		}

		if len(input) > 0 && input[0] == ':' {
			shell.Write([]byte("Unknown command: " + input + "\n"))
			continue
		}

		return input, nil, false
	}
}
