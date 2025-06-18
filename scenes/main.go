package scenes

import (
	"fmt"
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
		{"Help", "View available commands and navigation help", nil},
		{"Quit", "Exit the application", nil},
	}
}

func Main(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "")
	clearTerminal(shell)

	selectedIndex := 0

	shell.Write([]byte("🚀 Welcome to SSH Battle - Terminal Typing Game\n"))
	shell.Write([]byte("═══════════════════════════════════════════════\n\n"))
	shell.Write([]byte("Navigation:\n"))
	shell.Write([]byte("• Use ↑/↓ arrow keys or j/k to navigate\n"))
	shell.Write([]byte("• Press Enter to select\n"))
	shell.Write([]byte("• Type commands (like :help, :q, :game) anytime\n\n"))

	for {
		// Render menu
		renderMenu(shell, selectedIndex)

		input, err := readInput(s)
		if err != nil {
			s.Close()
			return nil
		}

		switch input {
		case "up", "k":
			if selectedIndex > 0 {
				selectedIndex--
			}
		case "down", "j":
			if selectedIndex < len(menuItems)-1 {
				selectedIndex++
			}
		case "enter":
			return handleMenuSelection(shell, s, p, selectedIndex)
		case "command":
			// Handle typed commands
			shell.Write([]byte("\n> "))
			line, nextScene, done := SafeReadInput(shell, s, p)
			if done {
				return nextScene
			}
			// If no scene change, continue with menu
			clearTerminal(shell)
			shell.Write([]byte("🚀 Welcome to SSH Battle - Terminal Typing Game\n"))
			shell.Write([]byte("═══════════════════════════════════════════════\n\n"))
			shell.Write([]byte("Navigation:\n"))
			shell.Write([]byte("• Use ↑/↓ arrow keys or j/k to navigate\n"))
			shell.Write([]byte("• Press Enter to select\n"))
			shell.Write([]byte("• Type commands (like :help, :q, :game) anytime\n\n"))

			// Show feedback for unknown input
			if line != "" {
				shell.Write(fmt.Appendf(nil, "Unknown input: %s\n\n", line))
			}
		}
	}
}

// Rest of the functions remain the same...
func renderMenu(shell *term.Terminal, selectedIndex int) {
	// Clear from cursor down
	shell.Write([]byte("\033[J"))

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
			// Show description for selected item
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

	// Handle arrow keys and other special sequences
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
	case input == "q" || input == "Q":
		return "command", nil
	case input == ":":
		return "command", nil
	case len(input) == 1 && (input[0] >= 32 && input[0] <= 126): // Printable ASCII
		return "command", nil
	case input == "\003": // Ctrl+C
		return "", fmt.Errorf("interrupted")
	default:
		// Ignore other inputs (like function keys, etc.)
		return "", nil
	}
}

func handleMenuSelection(shell *term.Terminal, s glider.Session, p *player.Player, selectedIndex int) Scene {
	selectedItem := menuItems[selectedIndex]

	switch selectedItem.Label {
	case "Quit":
		shell.Write([]byte("\nGoodbye! Thanks for playing SSH Battle! 👋\n"))
		s.Close()
		return nil
	default:
		// Navigate to the selected scene
		shell.Write([]byte(fmt.Sprintf("\n✨ Loading %s...\n", selectedItem.Label)))
		return selectedItem.Scene
	}
}
