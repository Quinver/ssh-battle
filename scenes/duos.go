package scenes

import (
	"context"
	"fmt"
	"log"
	"ssh-battle/player"
	"ssh-battle/util"
	"sync"
	"time"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

// Singleton duos behavior to ensure all players share the same game state
var duosBehaviorInstance *DuosRoomBehavior
var duosBehaviorOnce sync.Once

func getDuosBehavior() *DuosRoomBehavior {
	duosBehaviorOnce.Do(func() {
		duosBehaviorInstance = &DuosRoomBehavior{}
	})
	return duosBehaviorInstance
}

func Duos(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write([]byte("Welcome to Duos!\n"))
	shell.Write([]byte("Waiting for another player to join...\n\n"))

	// Get or create the duos room - use singleton behavior
	room := GetRoom("Duos", getDuosBehavior())
	
	// Join the room
	room.Join <- p
	defer func() { 
		room.Leave <- p 
		p.Ready = false // Reset ready state when leaving
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for incoming messages
	go func() {
		defer func() {
			// Recover from potential panic if channel is closed
			if r := recover(); r != nil {
				log.Printf("Message listener goroutine panic for %s: %v", p.Name, r)
			}
		}()
		
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-p.Messages:
				if !ok {
					return
				}
				shell.Write([]byte(msg + "\n"))
			}
		}
	}()

	// Announce player joined
	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s joined. Type 'ready' when you're ready to play!", p.Name)}

	// Wait for ready input
	for {
		shell.Write([]byte("Type 'ready' when you're ready to start (or use commands like :q, :lobby, etc.)\n"))
		input, nextScene, finished := SafeReadInput(shell, s, p)
		if finished {
			cancel()
			if nextScene != nil {
				return nextScene
			}
			return nil
		}

		if input == "ready" {
			p.Ready = true
			room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s is ready!", p.Name)}
			break
		} else if input != "" {
			shell.Write([]byte("Type 'ready' to start the game.\n"))
		}
	}

	// Wait for enough players and all to be ready
	shell.Write([]byte("Waiting for all players to be ready...\n"))
	for {
		room.mu.Lock()
		playerCount := len(room.Players)
		allReady := true
		readyCount := 0
		for _, player := range room.Players {
			if player.Ready {
				readyCount++
			} else {
				allReady = false
			}
		}
		room.mu.Unlock()

		if playerCount >= 2 && allReady {
			break
		}

		time.Sleep(500 * time.Millisecond)
		
		// Check if player wants to leave while waiting
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}

	// Only try to start the game once - let the room behavior handle it
	getDuosBehavior().TryStartGame(room)

	// Wait for game to actually start and get the sentence
	shell.Write([]byte("Preparing game...\n"))
	var sentence string
	duosBehavior := getDuosBehavior()
	for {
		duosBehavior.mu.Lock()
		started := duosBehavior.gameStarted
		sentence = duosBehavior.sentence
		duosBehavior.mu.Unlock()

		if started && sentence != "" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Countdown for all players
	shell.Write([]byte("\nGame starting in:\n"))
	for i := 3; i > 0; i-- {
		shell.Write(fmt.Appendf(nil, "%d...\n", i))
		time.Sleep(1 * time.Second)
	}
	shell.Write([]byte("GO!\n\n"))

	// Display the sentence
	green := "\033[32m"
	reset := "\033[0m"
	shell.Write(fmt.Appendf(nil, "%s%s%s\n\n", green, sentence, reset))
	shell.Write([]byte("Start typing:\n"))
	log.Printf("Player %s got sentence: %s", p.Name, sentence)

	// Record start time and get input
	start := time.Now()
	input, nextScene, finished := SafeReadInput(shell, s, p)
	elapsed := time.Since(start)

	if finished {
		cancel()
		if nextScene != nil {
			return nextScene
		}
		return nil
	}

	// Calculate and save score
	score := player.ScoreCalculation(sentence, input, elapsed)
	p.Scores = append(p.Scores, score)
	player.SaveScore(p.ID, score)

	// Display results
	shell.Write([]byte("\n=== Your Results ===\n"))
	fmt.Fprintf(s, "Accuracy: %.2f%%\n", *score.Accuracy)
	fmt.Fprintf(s, "WPM: %.1f\n", *score.WPM)
	fmt.Fprintf(s, "Time: %d seconds\n", *score.Duration)
	fmt.Fprintf(s, "TP Score: %.2f\n\n", *score.TP)

	// Announce completion to room
	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s finished! TP: %.2f", p.Name, *score.TP)}

	shell.Write([]byte("Press Enter to continue or use a command...\n"))
	_, nextScene, finished = SafeReadInput(shell, s, p)
	if finished {
		cancel()
		if nextScene != nil {
			return nextScene
		}
	}

	// Reset game state for next round
	getDuosBehavior().Reset()

	return Lobby
}

type DuosRoomBehavior struct {
	gameStarted bool
	sentence    string
	startTime   time.Time
	gameStarting bool  // New field to prevent race condition
	mu          sync.Mutex
}

func (d *DuosRoomBehavior) OnJoin(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s joined the duos room. (%d/2 players)", p.Name, len(r.Players))}
	log.Printf("%s joined the duos room.", p.Name)
}

func (d *DuosRoomBehavior) OnLeave(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s left the duos room. (%d players remaining)", p.Name, len(r.Players))}
	log.Printf("%s left the duos room.", p.Name)
	
	// Reset game if not enough players
	if len(r.Players) < 2 {
		d.Reset()
	}
}

func (d *DuosRoomBehavior) OnMessage(r *Room, msg RoomMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.Players {
		if p.Name != msg.Sender && p.Messages != nil {
			select {
			case p.Messages <- msg.Content:
			default:
				log.Printf("Dropping message for %s (channel full)", p.Name)
			}
		}
	}
}

// TryStartGame is called by each player, but only one will actually start the game
func (d *DuosRoomBehavior) TryStartGame(r *Room) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// If game is already started or starting, return
	if d.gameStarted || d.gameStarting {
		return
	}

	// Check if we have enough ready players
	r.mu.Lock()
	readyCount := 0
	totalPlayers := len(r.Players)
	for _, p := range r.Players {
		if p.Ready {
			readyCount++
		}
	}
	r.mu.Unlock()

	if totalPlayers < 2 || readyCount < totalPlayers {
		return // Not enough players or not all ready
	}

	// Mark as starting to prevent other players from starting
	d.gameStarting = true
	
	// Generate the sentence once
	d.sentence = util.GetSentences()
	
	// Mark as started
	d.gameStarted = true
	d.startTime = time.Now()

	log.Printf("Duos game started with sentence: %s", d.sentence)
	
	// Broadcast game start to all players
	r.Broadcast <- RoomMessage{"Server", "All players ready! Game starting..."}
}

func (d *DuosRoomBehavior) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gameStarted = false
	d.gameStarting = false
	d.sentence = ""
}
