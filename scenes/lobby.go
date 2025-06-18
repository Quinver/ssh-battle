package scenes

import (
	"fmt"
	"log"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func Lobby(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "")
	clearTerminal(shell)

	// Header
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 💬 \033[1;38;5;51mSSH Battle Lobby\033[0m\033[38;5;45m                           │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type messages to chat with other players\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Use commands like :q to quit or :help for more\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mChat:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))

	room := GetRoom("Lobby", LobbyRoomBehavior{})
	room.Join <- p
	defer func() {
		room.Leave <- p
	}()

	done := make(chan struct{})
	nextSceneCh := make(chan Scene, 1)

	// Receive message
	go func() {
		for {
			select {
			case msg, ok := <-p.Messages:
				if !ok {
					return
				}
				shell.Write([]byte("\033[38;5;252m\n" + msg + "\033[0m\n"))
				shell.Write([]byte("\033[38;5;208m> \033[0m"))
			case <-done:
				return
			}
		}
	}()

	// Send message
	go func() {
		for {
			line, nextScene, finished := SafeReadInput(shell, s, p)
			
			if finished {
				close(done)
				nextSceneCh <- nextScene
				return
			}
			if line != "" {
				room.Broadcast <- RoomMessage{
					Sender:  p.Name,
					Content: fmt.Sprintf("\033[38;5;51m[%s]\033[38;5;248m %s\033[0m", p.Name, line),
				}
			}
			shell.Write([]byte("\033[38;5;208m> \033[0m"))
		}
	}()

	// Keep room open
	for {
		select {
		case <-s.Context().Done():
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
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m%s joined the room.\033[0m", p.Name)}
	log.Printf("%s joined the lobby.", p.Name)
}

func (LobbyRoomBehavior) OnLeave(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;196m%s left the room.\033[0m", p.Name)}
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
