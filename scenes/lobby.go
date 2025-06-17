package scenes

import (
	"fmt"
	"log"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func Lobby(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write(fmt.Appendf(nil, "Welcome to the room: %s\n\n", "Lobby"))

	room := GetRoom("Lobby")
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


	// Keep room open
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

type LobbyRoomBehavior struct{}

func (LobbyRoomBehavior) OnJoin(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s joined the room.", p.Name)}
	log.Printf("%s joined the lobby.", p.Name)
}

func (LobbyRoomBehavior) OnLeave(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s left the room.", p.Name)}
	log.Printf("%s left the lobby.", p.Name)
}

func (LobbyRoomBehavior) OnMessage(r *Room, msg RoomMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.Players {
		if p.Name != msg.Sender {
			p.SendMessage(msg.Content)
		}
	}
}
