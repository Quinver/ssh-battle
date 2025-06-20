package scenes

import (
	"fmt"
	"log"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

type MenuItem struct {
	Label       string
	Description string
	Scene       Scene
}

var menuItems []MenuItem

func init() {
	menuItems = []MenuItem{
		{"Single Player Game", "Practice typing with randomly generated sentences", Game},
		{"Multiplayer Lobby", "Chat with other players and challenge them", Lobby},
		{"Duos Battle", "Real-time typing race with another player", Duos},
		{"Leaderboard", "View top scores from all players", Leaderboard},
		{"Your Scores", "View your personal typing history", ScoreList},
		{"Quit", "Exit the application", nil},
	}
}

func Main(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "")

	clearTerminal(shell)
	selectedIndex := 0

	// Initial render
	renderFullMenu(shell, selectedIndex)

	for {
		input, err := readInput(s)
		if err != nil {
			s.Close()
			return nil
		}

		switch input {
		case "up", "k":
			if selectedIndex > 0 {
				selectedIndex--
				renderFullMenu(shell, selectedIndex)
			}
		case "down", "j":
			if selectedIndex < len(menuItems)-1 {
				selectedIndex++
				renderFullMenu(shell, selectedIndex)
			}
		case "enter":
			return handleMenuSelection(shell, s, selectedIndex)
		case "command":
			// Handle typed commands
			shell.Write([]byte("\033[38;5;208m> \033[0m"))
			line, nextScene, done := SafeReadInput(shell, s, p)
			if done {
				return nextScene
			}

			// Re-render the full menu after command input
			renderFullMenu(shell, selectedIndex)

			// Show feedback for unknown input
			if line != "" {
				shell.Write(fmt.Appendf(nil, "\033[38;5;196mUnknown command: %s\033[0m\n\n", line))
			}
		}
	}
}

func renderFullMenu(shell *term.Terminal, selectedIndex int) {
	// Clear entire screen and move cursor to top
	clearTerminal(shell)

	// Header with gradient-like color
	shell.Write([]byte("\033[38;5;45m┌────────────────────────────────────────────────┐\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m│ 🚀 \033[1;38;5;51mSSH Battle - Terminal Typing Game\033[0m\033[38;5;45m         │\033[0m\n"))
	shell.Write([]byte("\033[38;5;45m└────────────────────────────────────────────────┘\033[0m\n\n"))

	// Instructions
	shell.Write([]byte("\033[38;5;229mNavigation Instructions:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m────────────────────────\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Use ↑/↓ or j/k to navigate\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Press Enter to select\033[0m\n"))
	shell.Write([]byte("\033[38;5;248m• Type commands (:help, :q, :game) anytime\033[0m\n\n"))

	// Menu items
	shell.Write([]byte("\033[38;5;229mSelect an Option:\033[0m\n"))
	shell.Write([]byte("\033[38;5;252m─────────────────\033[0m\n"))

	for i, item := range menuItems {
		var prefix string
		var style string
		var reset = "\033[0m"

		if i == selectedIndex {
			prefix = " \033[38;5;46m▶\033[0m "
			style = "\033[1;38;5;51m" // Bold cyan
		} else {
			prefix = "   "
			style = "\033[38;5;252m" // Light gray
		}

		shell.Write(fmt.Appendf(nil, "%s%s%-20s%s\n", style, prefix, item.Label, reset))

		if i == selectedIndex {
			shell.Write(fmt.Appendf(nil, "   \033[2;38;5;248m%s\033[0m\n", item.Description))
		}
	}

	shell.Write([]byte("\n"))
}

func readInput(s glider.Session) (string, error) {
	buffer := make([]byte, 10)
	n, err := s.Read(buffer)
	if err != nil {
		return "", err
	}

	input := string(buffer[:n])

	// Handle arrow keys and vim controls
	switch {
	case input == "\033[A": // Up arrow
		return "up", nil
	case input == "\033[B": // Down arrow
		return "down", nil
	case input == "\r" || input == "\n": // Enter
		return "enter", nil
	case input == "k" || input == "K":
		return "up", nil
	case input == "j" || input == "J":
		return "down", nil
	case input == ":":
		return "command", nil
	case len(input) == 1 && (input[0] >= 32 && input[0] <= 126): // Any other character are handles as a command input
		return "command", nil
	case input == "\003": // Ctrl+C
		return "", fmt.Errorf("interrupted")
	default:
		// Ignore other inputs (function keys, etc.)
		return "", nil
	}
}

func handleMenuSelection(shell *term.Terminal, s glider.Session, selectedIndex int) Scene {
	selectedItem := menuItems[selectedIndex]

	switch selectedItem.Label {
	case "Quit":
		shell.Write([]byte("\033[38;5;46m\nGoodbye! Thanks for playing SSH Battle! 👋\033[0m\n"))
		s.Close()
		return nil
	default:
		shell.Write(fmt.Appendf(nil, "\033[38;5;46m\n✨ Loading %s...\033[0m\n", selectedItem.Label))
		return selectedItem.Scene
	}
}

func SessionStart(s glider.Session) {
	p := player.GetOrCreatePlayer(s)
	if p == nil {
		log.Printf("Failed to create player for %s", s.User())
		s.Close()
		return
	}

	p.Session = s

	log.Printf("Player %s connected", p.Name)

	// Start with the main scene
	currentScene := Main

	// Scene loop - keep running until currentScene is nil
	for currentScene != nil {
		nextScene := currentScene(s, p)

		// Check if session is still active
		select {
		case <-s.Context().Done():
			log.Printf("Session context done for %s", p.Name)
			return
		default:
		}

		currentScene = nextScene
	}

	log.Printf("Player %s disconnected", p.Name)
}
