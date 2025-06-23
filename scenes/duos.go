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
		duosBehaviorInstance = &DuosRoomBehavior{
			gameTimeLimit: 60 * time.Second, // 60 second time limit
		}
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

	shell.Write([]byte("\033[38;5;229mControls:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type 'ready' to start the game\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :main to return to main menu\033[0m\n"))
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

	// Display the sentence with better formatting and time limit
	shell.Write([]byte("\033[38;5;229m📝 Type this sentence:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m" + strings.Repeat("─", 50) + "\033[0m\n"))
	shell.Write(fmt.Appendf(nil, "\033[1;38;5;252m%s\033[0m\n", sentence))
	shell.Write([]byte("\033[38;5;252m" + strings.Repeat("─", 50) + "\033[0m\n"))
	shell.Write([]byte("\033[38;5;196m⏰ Time limit: 60 seconds\033[0m\n\n"))
	shell.Write([]byte("\033[38;5;229m⌨️  Your typing:\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
	log.Printf("Player %s got sentence: %s", p.Name, sentence)

	// Record start time and get input with time limit
	start := time.Now()
	timeLimit := 60 * time.Second
	
	// Create a channel to receive input or timeout
	inputChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	// Start goroutine to read input
	go func() {
		input, _, finished := SafeReadInput(shell, s, p)
		if finished {
			errorChan <- fmt.Errorf("player disconnected")
			return
		}
		inputChan <- input
	}()
	
	var input string
	var timedOut bool
	
	// Wait for input or timeout
	select {
	case input = <-inputChan:
		// Got input
	case <-errorChan:
		// Player disconnected during game - they forfeit
		cancel()
		return nil
	case <-time.After(timeLimit):
		// Time limit exceeded
		timedOut = true
		input = "" // Empty input for timeout
		shell.Write([]byte("\033[2K\r")) // Clear current line
		shell.Write([]byte("\033[1;38;5;196m⏰ TIME'S UP! ⏰\033[0m\n"))
	}
	
	elapsed := min(time.Since(start), timeLimit)

	// Calculate and save score
	score := player.ScoreCalculation(sentence, input, elapsed)
	if timedOut {
		// Adjust score for timeout - set accuracy to 0 and low TP score
		zeroAccuracy := 0.0
		lowTP := 0.0
		score.Accuracy = &zeroAccuracy
		score.TP = &lowTP
	}
	p.Scores = append(p.Scores, score)
	player.SaveScore(p.ID, score)

	// Mark this player as finished and store their score
	duosBehavior.mu.Lock()
	duosBehavior.playerResults[p.Name] = PlayerResult{
		Player: p,
		Score:  &score,
		Input:  input,
		TimedOut: timedOut,
	}
	duosBehavior.mu.Unlock()

	// Only send completion message to OTHER players, not yourself
	if timedOut {
		// Don't broadcast timeout to self
		go func() {
			room.mu.Lock()
			for _, otherPlayer := range room.Players {
				if otherPlayer.Name != p.Name && otherPlayer.Messages != nil {
					select {
					case otherPlayer.Messages <- fmt.Sprintf("\033[38;5;196m⏰ %s ran out of time!\033[0m", p.Name):
					default:
					}
				}
			}
			room.mu.Unlock()
		}()
	} else {
		// Don't broadcast completion to self
		go func() {
			room.mu.Lock()
			for _, otherPlayer := range room.Players {
				if otherPlayer.Name != p.Name && otherPlayer.Messages != nil {
					select {
					case otherPlayer.Messages <- fmt.Sprintf("\033[38;5;46m🏁 %s finished typing!\033[0m", p.Name):
					default:
					}
				}
			}
			room.mu.Unlock()
		}()
	}

	// Show waiting message - player cannot exit during this phase
	clearTerminal(shell)
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ ⏳ \033[1;38;5;51mWAITING FOR OTHER PLAYER\033[0m\033[38;5;45m               │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))
	
	if timedOut {
		shell.Write([]byte("\033[38;5;196m⏰ You ran out of time!\033[0m\n\n"))
	} else {
		shell.Write([]byte("\033[38;5;46m✅ You finished typing!\033[0m\n\n"))
	}
	shell.Write([]byte("\033[38;5;248m🔒 Please wait for the other player to finish...\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m💡 You cannot exit until both players are done.\033[0m\n\n"))

	// Wait for both players to finish (or timeout)
	maxWaitTime := timeLimit + (10 * time.Second) // Extra time for the other player
	waitStart := time.Now()
	
	for {
		duosBehavior.mu.Lock()
		resultCount := len(duosBehavior.playerResults)
		allFinished := resultCount >= 2
		duosBehavior.mu.Unlock()
		
		if allFinished {
			break
		}
		
		// Check if we've waited too long (other player disconnected/timed out)
		if time.Since(waitStart) > maxWaitTime {
			shell.Write([]byte("\033[38;5;196m⚠️  Other player appears to have disconnected. Proceeding to results...\033[0m\n"))
			break
		}
		
		// Show periodic updates
		elapsed := time.Since(waitStart)
		if int(elapsed.Seconds()) % 5 == 0 {
			remaining := maxWaitTime - elapsed
			if remaining > 0 {
				shell.Write([]byte("\033[2K\r")) // Clear line
				shell.Write(fmt.Appendf(nil, "\033[38;5;248m⏳ Still waiting... (timeout in %.0f seconds)\033[0m\n", remaining.Seconds()))
			}
		}
		
		time.Sleep(1 * time.Second)
	}

	// Display results for both players
	clearTerminal(shell)
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 🏆 \033[1;38;5;51mFINAL BATTLE RESULTS\033[0m\033[38;5;45m                      │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Get all results and sort by TP score
	duosBehavior.mu.Lock()
	results := make([]PlayerResult, 0, len(duosBehavior.playerResults))
	for _, result := range duosBehavior.playerResults {
		results = append(results, result)
	}
	duosBehavior.mu.Unlock()

	// Sort results by TP score (descending)
	for i := range len(results)-1 {
		for j := i + 1; j < len(results); j++ {
			if *results[i].Score.TP < *results[j].Score.TP {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Display results
	for i, result := range results {
		rank := i + 1
		var rankIcon string
		switch rank {
		case 1:
			rankIcon = "🥇"
		case 2:
			rankIcon = "🥈"
		default:
			rankIcon = "🏅"
		}

		shell.Write(fmt.Appendf(nil, "\033[38;5;229m%s Rank %d: %s\033[0m\n", rankIcon, rank, result.Player.Name))
		shell.Write([]byte("\033[38;5;252m" + strings.Repeat("─", 30) + "\033[0m\n"))
		
		if result.TimedOut {
			shell.Write([]byte("\033[38;5;196m⏰ TIMED OUT\033[0m\n"))
		}
		
		shell.Write(fmt.Appendf(nil, "\033[38;5;248m🎯 Accuracy: \033[1;38;5;51m%.2f%%\033[0m\n", *result.Score.Accuracy))
		shell.Write(fmt.Appendf(nil, "\033[38;5;248m⚡ WPM: \033[1;38;5;51m%.1f\033[0m\n", *result.Score.WPM))
		shell.Write(fmt.Appendf(nil, "\033[38;5;248m⏱️  Time: \033[1;38;5;51m%d seconds\033[0m\n", *result.Score.Duration))
		shell.Write(fmt.Appendf(nil, "\033[38;5;248m🏆 TP Score: \033[1;38;5;51m%.2f\033[0m\n\n", *result.Score.TP))
	}

	// Send winner announcement only to other players  (Might change later)
	if len(results) >= 2 {
		winner := results[0]
		var winMessage string
		if *winner.Score.TP > *results[1].Score.TP {
			winMessage = fmt.Sprintf("\033[1;38;5;46m🎉 %s wins the battle! TP: %.2f 🎉\033[0m", 
				winner.Player.Name, *winner.Score.TP)
		} else {
			winMessage = "\033[1;38;5;248m🤝 It's a tie! Great battle! 🤝\033[0m"
		}
		
		// Send to other players only
		go func() {
			room.mu.Lock()
			for _, otherPlayer := range room.Players {
				if otherPlayer.Name != p.Name && otherPlayer.Messages != nil {
					select {
					case otherPlayer.Messages <- winMessage:
					default:
					}
				}
			}
			room.mu.Unlock()
		}()
	}

	shell.Write([]byte("\033[38;5;229mControls:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to return to lobby\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type :main to return to main menu\033[0m\n"))
	shell.Write([]byte("\033[38;5;208m> \033[0m"))
	
	_, nextScene, finished := SafeReadInput(shell, s, p)
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

type PlayerResult struct {
	Player   *player.Player
	Score    *player.Score
	Input    string
	TimedOut bool
}

type DuosRoomBehavior struct {
	gameStarted    bool
	sentence       string
	startTime      time.Time
	gameStarting   bool
	gameTimeLimit  time.Duration
	playerResults  map[string]PlayerResult
	mu             sync.Mutex
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
	
	// Send message to all players except the sender
	for _, p := range r.Players {
		// Skip the sender to avoid double messages
		if p.Name == msg.Sender {
			continue
		}
		
		if p.Messages != nil {
			select {
			case p.Messages <- msg.Content:
				// Message sent successfully
			case <-time.After(50 * time.Millisecond):
				// Timeout - channel might be full or blocked
				log.Printf("Message delivery timeout for %s (sender: %s)", p.Name, msg.Sender)
			default:
				// Channel full, skip this message
				log.Printf("Channel full for %s, skipping message from %s", p.Name, msg.Sender)
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
	d.playerResults = make(map[string]PlayerResult)

	log.Printf("Duos game started with %d players, sentence: %s", totalPlayers, d.sentence)
	r.Broadcast <- RoomMessage{"Server", "\033[1;38;5;46m🚀 All players ready! Battle commencing...\033[0m"}
}

func (d *DuosRoomBehavior) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gameStarted = false
	d.gameStarting = false
	d.sentence = ""
	d.playerResults = make(map[string]PlayerResult)
	log.Printf("Duos game state reset")
}
