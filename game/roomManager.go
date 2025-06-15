package game

import (
	"sync"

	"fmt"
	"log"
)

type Room struct {
	ID        string
	Players   map[string]*Player
	Broadcast chan string // channel to broadcast messages to all players
	Join      chan *Player
	Leave     chan *Player
	mu        sync.Mutex
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

var defaultRoomManager = &RoomManager{
	rooms: make(map[string]*Room),
}

func LobbyRoom(id string) *Room {
	return &Room{
		ID:        id,
		Players:   make(map[string]*Player),
		Broadcast: make(chan string),
		Join:      make(chan *Player),
		Leave:     make(chan *Player),
	}
}

func (r *Room) Run() {
	for {
		select {
		case p := <-r.Join:
			r.mu.Lock()
			r.Players[p.Name] = p
			r.mu.Unlock()
			r.Broadcast <- fmt.Sprintf("%s joined the room.", p.Name)
			log.Printf("%s joined the room.", p.Name)

		case p := <-r.Leave:
			r.mu.Lock()
			delete(r.Players, p.Name)
			close(p.Messages)
			r.mu.Unlock()
			r.Broadcast <- fmt.Sprintf("%s left the room.", p.Name)
			log.Printf("%s left the room.", p.Name)

		case msg := <-r.Broadcast:
			r.mu.Lock()
			for _, player := range r.Players {
				player.SendMessage(msg)
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
			Broadcast: make(chan string, 10),
			Join:      make(chan *Player, 10),
			Leave:     make(chan *Player, 10),
		}
		go room.Run()
		defaultRoomManager.rooms[id] = room
	}

	return room
}
