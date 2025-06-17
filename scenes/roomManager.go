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
			// Always create a new message channel when joining a room
			// This handles the case where the player had a closed channel from a previous room
			if p.Messages != nil {
				// Close the old channel if it exists (it might already be closed, but that's safe)
				select {
				case <-p.Messages:
					// Channel was already closed, do nothing
				default:
					close(p.Messages)
				}
			}
			p.Messages = make(chan string, 10)
			
			r.mu.Lock()
			r.Players[p.Name] = p
			r.mu.Unlock()
			r.Behavior.OnJoin(r, p)

		case p := <-r.Leave:
			r.mu.Lock()
			delete(r.Players, p.Name)
			// Close the message channel when leaving
			if p.Messages != nil {
				close(p.Messages)
				p.Messages = nil
			}
			empty := len(r.Players) == 0
			r.mu.Unlock()
			r.Behavior.OnLeave(r, p)

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

func (b *DuosRoomBehavior) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.gameStarted = false
	b.sentence = ""
}
