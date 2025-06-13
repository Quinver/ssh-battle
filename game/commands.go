// Helper file for global commands
package game

import "golang.org/x/term"

func HandleCommands(input string, shell *term.Terminal) (handled bool, quit bool) {
	if len(input) > 0 && input[0] == ':' {
		switch input {
		case ":q":
			shell.Write([]byte("Goodbye!\n"))
			return true, true // handled and quit
		case ":help":
			shell.Write([]byte("Commands:\n:q - quit\n:help - show this help\n"))
			return true, false // handled, but don't quit
		default:
			shell.Write([]byte("Unknown command: " + input + "\n"))
			return true, false
		}
	}
	return false, false
}

func ReadInput(shell *term.Terminal) (string, bool) {
	for {
		input, err := shell.ReadLine()
		if err != nil {
			return "", true // quit on error
		}

		handled, quit := HandleCommands(input, shell)
		if handled {
			if quit {
				return "", true
			}
			continue // command handled but no quit, ask for input again
		}

		return input, false
	}
}
