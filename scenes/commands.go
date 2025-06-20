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
			Description: "quit the application",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   nil,
			Quit:        true,
		},
		":help": {
			Description: "show this help menu",
			Handler:     helpHandler,
			NextScene:   nil,
		},
		":main": {
			Description: "go to main menu",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Main,
		},
		":game": {
			Description: "start single player game",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Game,
		},
		":lobby": {
			Description: "go to multiplayer lobby",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Lobby,
		},
		":scores": {
			Description: "view your score history",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   ScoreList,
		},
		":leaderboard": {
			Description: "view global leaderboard",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Leaderboard,
		},
		":duos": {
			Description: "join duos battle arena",
			Handler:     func(_ *term.Terminal) {},
			NextScene:   Duos,
		},
	}

	AddAlias(":exit", ":q")
	AddAlias(":quit", ":q")
	AddAlias(":home", ":main")
	AddAlias(":menu", ":main")
	AddAlias(":single", ":game")
	AddAlias(":multi", ":lobby")
	AddAlias(":chat", ":lobby")
	AddAlias(":top", ":leaderboard")
	AddAlias(":history", ":scores")
	AddAlias(":battle", ":duos")
}

// Enhanced help command with better formatting
func helpHandler(shell *term.Terminal) {
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 📚 \033[1;38;5;51mSSH Battle Commands\033[0m\033[38;5;45m                        │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mNavigation:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m───────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press \033[1;38;5;51mESC\033[0m\033[38;5;248m to return to main menu from any scene\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type commands starting with \033[1;38;5;51m:\033[0m\033[38;5;248m (colon)\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mAvailable Commands:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────────\033[0m\n"))

	mainCommands := make([]string, 0, len(commandRegistry))
	for cmd := range commandRegistry {
		if _, isAlias := aliases[cmd]; !isAlias {
			mainCommands = append(mainCommands, cmd)
		}
	}

	sort.Strings(mainCommands)

	for _, cmd := range mainCommands {
		data := commandRegistry[cmd]
		shell.Write(fmt.Appendf(nil, "\033[1;38;5;51m%-15s\033[0m \033[38;5;248m%s\033[0m\n", cmd, data.Description))

		// Show aliases for this command
		aliases_for_cmd := make([]string, 0)
		for alias, original := range aliases {
			if original == cmd {
				aliases_for_cmd = append(aliases_for_cmd, alias)
			}
		}
		if len(aliases_for_cmd) > 0 {
			shell.Write(fmt.Appendf(nil, "\033[38;5;240m                Aliases: %s\033[0m\n", 
				fmt.Sprintf("%v", aliases_for_cmd)))
		}
	}

	shell.Write([]byte("\n\033[38;5;229mExamples:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type \033[1;38;5;51m:game\033[0m\033[38;5;248m to start single player\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type \033[1;38;5;51m:duos\033[0m\033[38;5;248m to join battle arena\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press \033[1;38;5;51mESC\033[0m\033[38;5;248m for quick main menu access\033[0m\n\n"))
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
			shell.Write([]byte("\033[38;5;196m❌ Unknown command: " + input + "\033[0m\n"))
			shell.Write([]byte("\033[38;5;248mType \033[1;38;5;51m:help\033[0m\033[38;5;248m for available commands\033[0m\n"))
			continue
		}

		return input, nil, false
	}
}

// Helper function to show control hints in any scene
func ShowControlHints(shell *term.Terminal, customHints ...string) {
	shell.Write([]byte("\033[38;5;229mControls:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	
	// Show custom hints first
	for _, hint := range customHints {
		shell.Write([]byte("\033[38;5;248m• " + hint + "\033[0m\n"))
	}
	
	// Always show these universal controls
	shell.Write([]byte("\033[38;5;248m• Press \033[1;38;5;51mESC\033[0m\033[38;5;248m to return to main menu\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type \033[1;38;5;51m:help\033[0m\033[38;5;248m for all commands\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type \033[1;38;5;51m:q\033[0m\033[38;5;248m to quit\033[0m\n\n"))
}
