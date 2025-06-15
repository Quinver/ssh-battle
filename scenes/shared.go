package scenes

import (
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

// Scene function type for switching between scenes
type Scene func(glider.Session, *player.Player) Scene

func clearTerminal(shell *term.Terminal) {
	shell.Write([]byte("\033[2J\033[H")) // ANSI escape to clear screen and move cursor home
}
