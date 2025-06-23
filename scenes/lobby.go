package scenes

import (
	"fmt"
	"log"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
)

func Lobby(s glider.Session, p *player.Player) Scene {
	shell := p.Shell
	clearTerminal(shell)

	// Add the missing header and UI like other scenes
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 💬 \033[1;38;5;51mMultiplayer Lobby\033[0m\033[38;5;45m                        │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type messages to chat with other players\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :main to return to main menu\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :q to quit or :help for more commands\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mChat:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────\033[0m\n"))

	// Show the initial prompt
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
				// Clear current line, print message, then restore prompt
				shell.Write([]byte("\033[2K\r")) // Clear line
				shell.Write([]byte(msg + "\n"))
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

			// finished means going to next scene or quiting
			if finished {
				close(done)
				nextSceneCh <- nextScene
				return
			}
			room.Broadcast <- RoomMessage{
				Sender:  p.Name,
				Content: fmt.Sprintf("[%s] %s", p.Name, line),
			}
			// Show prompt again after sending message
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
