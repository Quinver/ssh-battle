package scenes

import (
	"context"
	"fmt"
	"log"
	"ssh-battle/player"
	"ssh-battle/util"
	"strings"
	"sync"
	"time"

	glider "github.com/gliderlabs/ssh"
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
	shell := p.Shell
	clearTerminal(shell)

	// Header
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ ⚔️ \033[1;38;5;51mDuos Typing Battle\033[0m\033[38;5;45m                        │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Enhanced Instructions with Escape key info
	shell.Write([]byte("\033[38;5;229mControls:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type 'ready' to start the game\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press ESC to return to main menu\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Use :q to quit, :help for all commands\033[0m\n\n"))

	shell.Write([]byte("\033[38;5;229mWaiting:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248mWaiting for another player to join...\033[0m\n\n"))

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

	// Listen for incoming messages with improved error handling
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
				// Clear current line and print message, then restore prompt
				shell.Write([]byte("\033[2K\r")) // Clear line
				shell.Write([]byte("\033[38;5;252m" + msg + "\033[0m\n"))
				shell.Write([]byte("\033[38;5;208m> \033[0m"))
			}
		}
	}()

	// Announce player joined with better formatting
	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m🎮 %s joined the battle! (%d/2 players)\033[0m", p.Name, len(room.Players)+1)}

	// Wait for ready input with enhanced input handling
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
			room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m⚡ %s is ready to battle!\033[0m", p.Name)}
			break
		} else if input != "" {
			shell.Write([]byte("\033[38;5;196m❌ Type 'ready' to start the game or ESC for main menu.\033[0m\n"))
		}
	}

	// Wait for enough players and all to be ready with better status updates
	shell.Write([]byte("\033[38;5;248m⏳ Waiting for all players to be ready...\033[0m\n"))
	lastStatus := ""
	for {
		room.mu.Lock()
		playerCount := len(room.Players)
		readyCount := 0
		playerNames := make([]string, 0, len(room.Players))
		for _, player := range room.Players {
			playerNames = append(playerNames, player.Name)
			if player.Ready {
				readyCount++
			}
		}
		room.mu.Unlock()

		// Update status if changed
		currentStatus := fmt.Sprintf("Players: %d/2 | Ready: %d/%d", playerCount, readyCount, playerCount)
		if currentStatus != lastStatus {
			shell.Write([]byte("\033[2K\r")) // Clear line
			shell.Write([]byte("\033[38;5;248m" + currentStatus + "\033[0m\n"))
			lastStatus = currentStatus
		}

		if playerCount >= 2 && readyCount == playerCount {
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
	shell.Write([]byte("\033[38;5;248m🎮 Preparing battle arena...\033[0m\n"))
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

	// Enhanced countdown for all players
	clearTerminal(shell)
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ ⚔️ \033[1;38;5;51mBATTLE STARTING\033[0m\033[38;5;45m                             │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))
	
	for i := 3; i > 0; i-- {
		shell.Write([]byte("\033[2K\r")) // Clear line
		shell.Write(fmt.Appendf(nil, "\033[38;5;46m🚀 Starting in %d...\033[0m", i))
		time.Sleep(1 * time.Second)
	}
	shell.Write([]byte("\033[2K\r")) // Clear line
	shell.Write([]byte("\033[1;38;5;46m⚡ GO! GO! GO! ⚡\033[0m\n\n"))

	// Display the sentence with better formatting
	shell.Write([]byte("\033[38;5;229m📝 Type this sentence:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m" + strings.Repeat("─", 50) + "\033[0m\n"))
	shell.Write(fmt.Appendf(nil, "\033[1;38;5;252m%s\033[0m\n", sentence))
	shell.Write([]byte("\033[38;5;252m" + strings.Repeat("─", 50) + "\033[0m\n\n"))
	shell.Write([]byte("\033[38;5;229m⌨️  Your typing:\033[0m\n"))
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

	// Display results with better formatting
	clearTerminal(shell)
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 📊 \033[1;38;5;51mYOUR BATTLE RESULTS\033[0m\033[38;5;45m                       │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))
	
	shell.Write(fmt.Appendf(nil, "\033[38;5;248m🎯 Accuracy: \033[1;38;5;51m%.2f%%\033[0m\n", *score.Accuracy))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248m⚡ WPM: \033[1;38;5;51m%.1f\033[0m\n", *score.WPM))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248m⏱️  Time: \033[1;38;5;51m%d seconds\033[0m\n", *score.Duration))
	shell.Write(fmt.Appendf(nil, "\033[38;5;248m🏆 TP Score: \033[1;38;5;51m%.2f\033[0m\n\n", *score.TP))

	// Enhanced completion announcement to room
	room.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m🏁 %s finished! TP: %.2f | Accuracy: %.1f%% | WPM: %.1f\033[0m", 
		p.Name, *score.TP, *score.Accuracy, *score.WPM)}

	shell.Write([]byte("\033[38;5;229mControls:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to return to lobby\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press ESC to return to main menu\033[0m\n"))
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
	playerCount := len(r.Players)
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;46m🎮 %s joined the duos arena! (%d/2 players)\033[0m", p.Name, playerCount)}
	log.Printf("%s joined the duos room. Total players: %d", p.Name, playerCount)
}

func (d *DuosRoomBehavior) OnLeave(r *Room, p *player.Player) {
	playerCount := len(r.Players)
	r.Broadcast <- RoomMessage{"Server", fmt.Sprintf("\033[38;5;196m👋 %s left the duos arena. (%d players remaining)\033[0m", p.Name, playerCount)}
	log.Printf("%s left the duos room. Remaining players: %d", p.Name, playerCount)

	// Reset game if not enough players
	if playerCount < 2 {
		d.Reset()
		if playerCount > 0 {
			r.Broadcast <- RoomMessage{"Server", "\033[38;5;248m🔄 Game reset - waiting for another player...\033[0m"}
		}
	}
}

func (d *DuosRoomBehavior) OnMessage(r *Room, msg RoomMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Improved message delivery with better error handling
	for _, p := range r.Players {
		if p.Name != msg.Sender && p.Messages != nil {
			select {
			case p.Messages <- msg.Content:
				// Message sent successfully
			case <-time.After(100 * time.Millisecond):
				// Timeout - channel might be full or blocked
				log.Printf("Message delivery timeout for %s (sender: %s)", p.Name, msg.Sender)
				// Try to drain one message and retry
				select {
				case <-p.Messages:
					select {
					case p.Messages <- msg.Content:
					default:
						log.Printf("Still couldn't deliver message to %s after drain", p.Name)
					}
				default:
				}
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

	log.Printf("Duos game started with %d players, sentence: %s", totalPlayers, d.sentence)
	r.Broadcast <- RoomMessage{"Server", "\033[1;38;5;46m🚀 All players ready! Battle commencing...\033[0m"}
}

func (d *DuosRoomBehavior) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gameStarted = false
	d.gameStarting = false
	d.sentence = ""
	log.Printf("Duos game state reset")
}
