// Helper file for global commands
package scenes

import (
	"fmt"
	"sort"
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

var aliases = map[string]string{}

func init() {
	commandRegistry = map[string]Command{
		":q": {
			Description: "quit",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   nil,
			Quit:        true,
		},
		":help": {
			Description: "show this help",
			Handler:     helpHandler,
			NextScene:   nil,
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
			NextScene:   Lobby,
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

	AddAlias(":exit", ":q")
	AddAlias(":quit", ":q")
	AddAlias(":home", ":main")
}

// Make help command look pretty ✨
func helpHandler(shell *term.Terminal) {
	shell.Write([]byte("Commands:\n"))

	mainCommands := make([]string, 0, len(commandRegistry))
	for cmd := range commandRegistry {
		if _, isAlias := aliases[cmd]; !isAlias {
			mainCommands = append(mainCommands, cmd)
		}
	}

	sort.Strings(mainCommands)

	for _, cmd := range mainCommands {
		data := commandRegistry[cmd]
		shell.Write(fmt.Appendf(nil, "%s - %s\n", cmd, data.Description))

		for alias, original := range aliases {
			if original == cmd {
				shell.Write(fmt.Appendf(nil, "  %s\n", alias))
			}
		}
	}
}

func AddAlias(alias, original string) {
	aliases[alias] = original
	if cmd, ok := commandRegistry[original]; ok {
		commandRegistry[alias] = cmd
	}
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
			if cmd.Quit {
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
