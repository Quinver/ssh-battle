package game

import (
	"sync"

	"fmt"
	"log"
)

type Room struct {
	ID        string
	Players   map[string]*Player
	Broadcast chan RoomMessage // channel to broadcast messages to all players
	Join      chan *Player
	Leave     chan *Player
	mu        sync.Mutex
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

type RoomMessage struct {
	Sender  string
	Content string
}

var defaultRoomManager = &RoomManager{
	rooms: make(map[string]*Room),
}

func (r *Room) Run() {
	for {
		select {
		case p := <-r.Join:
			r.mu.Lock()
			r.Players[p.Name] = p
			r.mu.Unlock()
			// system messages don’t have a sender
			r.Broadcast <- RoomMessage{
				Sender:  "Server",
				Content: fmt.Sprintf("%s joined the room.", p.Name),
			}
			log.Printf("%s joined the room.", p.Name)

		case p := <-r.Leave:
			r.mu.Lock()
			delete(r.Players, p.Name)
			close(p.Messages)
			r.mu.Unlock()
			r.Broadcast <- RoomMessage{
				Sender:  "Server",
				Content: fmt.Sprintf("%s left the room.", p.Name),
			}
			log.Printf("%s left the room.", p.Name)

		case msg := <-r.Broadcast:
			r.mu.Lock()
			for _, player := range r.Players {
				if player.Name != msg.Sender {
					player.SendMessage(msg.Content)
				}
			}
			r.mu.Unlock()
		}
	}
}

func (p *Player) SendMessage(msg string) {
	if p == nil {
		return
	}
	select {
	case p.Messages <- msg:
	default:
		log.Printf("Dropping message for %s (channel full)", p.Name)
	}
}

// GetRoom returns an existing room or creates a new one if it doesn't exist.
func GetRoom(id string) *Room {
	defaultRoomManager.mu.Lock()
	defer defaultRoomManager.mu.Unlock()

	room, exists := defaultRoomManager.rooms[id]
	if !exists {
		room = &Room{
			ID:        id,
			Players:   make(map[string]*Player),
			Broadcast: make(chan RoomMessage, 10), // buffered to reduce blocking
			Join:      make(chan *Player, 10),
			Leave:     make(chan *Player, 10),
		}
		defaultRoomManager.rooms[id] = room
		go room.Run() // start once per room here
	}

	return room
}
