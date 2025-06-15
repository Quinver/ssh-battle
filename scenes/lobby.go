package scenes

import (
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
)

func MultiplayerLobby(s glider.Session, p *player.Player) Scene {
	return MultiplayerRoom(s, p, "main-lobby")
}
