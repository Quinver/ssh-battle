package scenes

import (
	"ssh-battle/player"
	"sync"
)

type Room struct {
	ID        string
	Players   map[string]*player.Player
	Broadcast chan RoomMessage
	Join      chan *player.Player
	Leave     chan *player.Player
	Behavior  RoomBehavior
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

type RoomBehavior interface {
	OnJoin(r *Room, p *player.Player)
	OnLeave(r *Room, p *player.Player)
	OnMessage(r *Room, msg RoomMessage)
}

type ResettableBehavior interface {
	RoomBehavior
	Reset()
}

var defaultRoomManager = &RoomManager{
	rooms: make(map[string]*Room),
}

func (r *Room) Run() {
	for {
		select {
		case p := <-r.Join:
			// Ensure the player has a fresh message channel
			// Don't close the old channel here as it might be in use by other goroutines
			p.Messages = make(chan string, 10)
			
			r.mu.Lock()
			r.Players[p.Name] = p
			r.mu.Unlock()
			r.Behavior.OnJoin(r, p)

		case p := <-r.Leave:
			r.mu.Lock()
			delete(r.Players, p.Name)
			empty := len(r.Players) == 0
			r.mu.Unlock()
			r.Behavior.OnLeave(r, p)

			// Don't close the message channel here - let the scene handle it
			// The channel might still be in use by the scene's goroutines

			if empty {
				defaultRoomManager.mu.Lock()
				delete(defaultRoomManager.rooms, r.ID)
				defaultRoomManager.mu.Unlock()

				// Exit the Run goroutine
				return
			}

		case msg := <-r.Broadcast:
			r.Behavior.OnMessage(r, msg)
		}
	}
}

// GetRoom returns an existing room or creates a new one if it doesn't exist.
func GetRoom(id string, behavior RoomBehavior) *Room {
	defaultRoomManager.mu.Lock()
	defer defaultRoomManager.mu.Unlock()

	room, exists := defaultRoomManager.rooms[id]
	if !exists {
		room = &Room{
			ID:        id,
			Players:   make(map[string]*player.Player),
			Broadcast: make(chan RoomMessage, 10), // buffered to reduce blocking
			Join:      make(chan *player.Player, 10),
			Leave:     make(chan *player.Player, 10),
			Behavior:  behavior,
		}
		defaultRoomManager.rooms[id] = room
		go room.Run() // start once per room here

		if resettable, ok := behavior.(ResettableBehavior); ok {
			resettable.Reset()
		}
	}

	return room
}
