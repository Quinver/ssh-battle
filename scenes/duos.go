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
	shell := term.NewTerminal(s, "")
	clearTerminal(shell)

	// Header
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ ⚔️ \033[1;38;5;51mDuos Typing Battle\033[0m\033[38;5;45m                        │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mInstructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type 'ready' to start the game\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Use commands like :q or :help for more\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mWaiting:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248mWaiting for another player to join...\033[0m\n"))

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
				shell.Write([]byte("\033[38;5;252m\n" + msg + "\033[0m\n"))
				shell.Write([]byte("\033[38;5;208m> \033[0m"))
			}
		}
	}()

	// Announce player joined
	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m%s joined. Type 'ready' when you're ready to play!\033[0m", p.Name)}

	// Wait for ready input
	for {
		shell.Write([]byte("\033[38;5;46mType 'ready' when you're ready to start...\033[0m\n"))
		shell.Write([]byte("\033[38;5;208m> \033[0m"))
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
			room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m%s is ready!\033[0m", p.Name)}
			break
		} else if input != "" {
			shell.Write([]byte("\033[38;5;196mType 'ready' to start the game.\033[0m\n"))
		}
	}

	// Wait for enough players and all to be ready
	shell.Write([]byte("\033[38;5;248mWaiting for all players to be ready...\033[0m\n"))
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
	shell.Write([]byte("\033[38;5;248mPreparing game...\033[0m\n"))
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
	shell.Write([]byte("\033[38;5;229m\nGame Starting:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────────\033[0m\n"))
	for i := 3; i > 0; i-- {
		shell.Write(fmt.Appendf(nil, "\033[38;5;46m%d...\033[0m\n", i))
		time.Sleep(1 * time.Second)
	}
	shell.Write([]byte("\033[38;5;46mGO!\033[0m\n\n"))

	// Display the sentence
	shell.Write([]byte("\033[38;5;229mSentence:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	shell.Write(fmt.Appendf(nil, "\033[38;5;252m%s\033[0m\n\n", sentence))
	shell.Write([]byte("\033[38;5;229mType Here:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m──────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
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
	shell.Write([]byte("\033[38;5;229m\nYour Results:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────────\033[0m\n"))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mAccuracy: \033[38;5;51m%.2f%%\033[0m\n", *score.Accuracy))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mWPM: \033[38;5;51m%.1f\033[0m\n", *score.WPM))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mTime: \033[38;5;51m%d seconds\033[0m\n", *score.Duration))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248mTP Score: \033[38;5;51m%.2f\033[0m\n\n", *score.TP))

	// Announce completion to room
	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m%s finished! TP: %.2f\033[0m", p.Name, *score.TP)}

	shell.Write([]byte("\033[38;5;46mPress Enter to return to lobby...\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
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
	gameStarted  bool
	sentence     string
	startTime    time.Time
	gameStarting bool
	mu           sync.Mutex
}

func (d *DuosRoomBehavior) OnJoin(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m%s joined the duos room. (%d/2 players)\033[0m", p.Name, len(r.Players))}
	log.Printf("%s joined the duos room.", p.Name)
}

func (d *DuosRoomBehavior) OnLeave(r *Room, p *player.Player) {
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;196m%s left the duos room. (%d players remaining)\033[0m", p.Name, len(r.Players))}
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

func (d *DuosRoomBehavior) TryStartGame(r *Room) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.gameStarted || d.gameStarting {
		return
	}

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
		return
	}

	d.gameStarting = true
	d.sentence = util.GetSentences()
	d.gameStarted = true
	d.startTime = time.Now()

	log.Printf("Duos game started with sentence: %s", d.sentence)
	r.Broadcast <- RoomMessage{"Server", "\033[38;5;46mAll players ready! Game starting...\033[0m"}
}

func (d *DuosRoomBehavior) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gameStarted = false
	d.gameStarting = false
	d.sentence = ""
}
