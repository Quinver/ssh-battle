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

var duosRoomBehavior = &DuosRoomBehavior{}

func Duos(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	duosBehavior := &DuosRoomBehavior{}
	room := GetRoom("Duos", duosBehavior)
	
	// Ensure player has a fresh message channel
	if p.Messages == nil {
		p.Messages = make(chan string, 10)
	}
	
	room.Join <- p
	defer func() { 
		room.Leave <- p
		// Don't close the channel here - let the room manager handle it
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start message listener goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Message listener panic for %s: %v", p.Name, r)
			}
		}()
		listenForMessages(ctx, p, shell)
	}()

	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s is getting ready. Type 'ready' to begin.", p.Name)}

	// Ready phase
	for {
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
			room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s is ready.", p.Name)}
			break
		}
		shell.Write([]byte("Type 'ready' when you're ready to start.\n"))
	}

	startGameIfFirst(room)
	waitForGameStart()

	// Game input phase
	input, nextScene, finished := SafeReadInput(shell, s, p)
	if finished {
		cancel()
		if nextScene != nil {
			return nextScene
		}
		return nil
	}

	score := handleScoring(p, input)
	showScore(s, score)

	// Final input phase
	_, nextScene, finished = SafeReadInput(shell, s, p)
	if finished {
		cancel()
		if nextScene != nil {
			return nextScene
		}
		return nil
	}

	cancel()
	return Lobby
}

type DuosRoomBehavior struct {
	gameStarted bool
	sentence    string
	mu          sync.Mutex
}

func (d *DuosRoomBehavior) OnJoin(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s joined the room.", p.Name)}
	log.Printf("%s joined the duos room.", p.Name)
}

func (d *DuosRoomBehavior) OnLeave(r *Room, p *player.Player) {
	// Reset player ready state when leaving
	p.Ready = false
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("%s left the room.", p.Name)}
	log.Printf("%s left the duos room.", p.Name)
}

func (d *DuosRoomBehavior) OnMessage(r *Room, msg RoomMessage) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Send message to all players except sender
	for _, p := range r.Players {
		if p.Name != msg.Sender {
			// Use safe message sending
			select {
			case p.Messages <- msg.Content:
			default:
				log.Printf("Warning: Message channel full for player %s, dropping message", p.Name)
			}
		}
	}
}

func (r *Room) WaitForAllReady() {
	for {
		r.mu.Lock()
		allReady := true
		playerCount := len(r.Players)
		
		if playerCount == 0 {
			r.mu.Unlock()
			return
		}
		
		for _, p := range r.Players {
			if !p.Ready {
				allReady = false
				break
			}
		}
		r.mu.Unlock()

		if allReady && playerCount > 0 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func listenForMessages(ctx context.Context, p *player.Player, shell *term.Terminal) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-p.Messages:
			if !ok {
				// Channel closed, exit gracefully
				return
			}
			// Use safe write to terminal
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Terminal write panic for %s: %v", p.Name, r)
					}
				}()
				shell.Write([]byte(msg + "\n"))
			}()
		}
	}
}

func startGameIfFirst(room *Room) {
	duosRoomBehavior.mu.Lock()
	defer duosRoomBehavior.mu.Unlock()

	if duosRoomBehavior.gameStarted {
		return
	}
	duosRoomBehavior.gameStarted = true

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Game start panic: %v", r)
			}
		}()
		
		room.WaitForAllReady()

		duosRoomBehavior.mu.Lock()
		duosRoomBehavior.sentence = util.GetSentences()
		log.Printf("Sentence generated: %s", duosRoomBehavior.sentence)
		green := "\033[32m"
		reset := "\033[0m"
		room.Broadcast <- RoomMessage{"Server", green + duosRoomBehavior.sentence + reset}
		duosRoomBehavior.mu.Unlock()
	}()
}

func waitForGameStart() {
	for {
		duosRoomBehavior.mu.Lock()
		started := duosRoomBehavior.gameStarted
		duosRoomBehavior.mu.Unlock()
		if started {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func handleScoring(p *player.Player, input string) *player.Score {
	duosRoomBehavior.mu.Lock()
	sentence := duosRoomBehavior.sentence
	duosRoomBehavior.mu.Unlock()

	elapsed := time.Since(time.Now().Add(-time.Second)) // simulate start time
	score := player.ScoreCalculation(sentence, input, elapsed)
	p.Scores = append(p.Scores, score)
	player.SaveScore(p.ID, score)

	return &score
}

func showScore(s glider.Session, score *player.Score) {
	fmt.Fprintf(s, "Accuracy: %.2f%%\nWPM: %.1f\nTime: %d\nTP: %.2f\n\n",
		*score.Accuracy, *score.WPM, *score.Duration, *score.TP)
}
