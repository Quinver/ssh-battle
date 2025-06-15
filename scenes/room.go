package scenes

import (
	"fmt"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func MultiplayerRoom(s glider.Session, p *player.Player, roomID string) Scene {
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

