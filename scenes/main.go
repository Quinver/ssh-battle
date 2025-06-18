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
			shell.Write([]byte("\n> "))
			line, nextScene, done := SafeReadInput(shell, s, p)
			if done {
				return nextScene
			}
			
			// Re-render the full menu after command input
			renderFullMenu(shell, selectedIndex)

			// Show feedback for unknown input
			if line != "" {
				shell.Write(fmt.Appendf(nil, "Unknown input: %s\n\n", line))
			}
		}
	}
}

func renderFullMenu(shell *term.Terminal, selectedIndex int) {
	// Clear entire screen and move cursor to top
	clearTerminal(shell)
	
	shell.Write([]byte("🚀 Welcome to SSH Battle - Terminal Typing Game\n"))
	shell.Write([]byte("═══════════════════════════════════════════════\n\n"))
	shell.Write([]byte("Navigation:\n"))
	shell.Write([]byte("• Use ↑/↓ arrow keys or j/k to navigate\n"))
	shell.Write([]byte("• Press Enter to select\n"))
	shell.Write([]byte("• Type commands (like :help, :q, :game) anytime\n\n"))
	
	// Render menu items
	shell.Write([]byte("Select an option:\n"))
	shell.Write([]byte("─────────────────\n"))

	for i, item := range menuItems {
		var prefix string
		var style string
		var reset = "\033[0m"

		if i == selectedIndex {
			prefix = "► "
			style = "\033[1;36m" // Bold cyan
		} else {
			prefix = "  "
			style = "\033[37m" // Light gray
		}

		shell.Write(fmt.Appendf(nil, "%s%s%s%s\n", style, prefix, item.Label, reset))

		if i == selectedIndex {
			shell.Write(fmt.Appendf(nil, "  \033[2;37m%s\033[0m\n", item.Description))
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
		shell.Write([]byte("\nGoodbye! Thanks for playing SSH Battle! 👋\n"))
		s.Close()
		return nil
	default:
		shell.Write(fmt.Appendf(nil, "\n✨ Loading %s...\n", selectedItem.Label))
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
