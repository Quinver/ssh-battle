package server

import (
	"log"
	"sync"

	"ssh-battle/game"
	"ssh-battle/keys"

	glider "github.com/gliderlabs/ssh"
)

var loggedInUsers = make(map[string]bool)
var loggedInMu sync.Mutex

func StartServer() {
	hostKey, err := keys.LoadHostKey("host_key.pem")
	if err != nil {
		log.Fatal("Failed to load host key:", err)
	}

	server := &glider.Server{
		Addr: ":2222",
		PasswordHandler: func(ctx glider.Context, password string) bool {
			return game.CheckPassword(ctx.User(), password)
		},
		Handler: func(s glider.Session) {
			username := s.User()

			// Check if user already logged in
			loggedInMu.Lock()
			if loggedInUsers[username] {
				loggedInMu.Unlock()
				s.Write([]byte("User already logged in elsewhere. Disconnecting...\n"))
				s.Close()
				return
			}
			loggedInUsers[username] = true
			loggedInMu.Unlock()
			
			// Delete user from currently save users after session ends
			defer func() {
				loggedInMu.Lock()
				delete(loggedInUsers, username)
				loggedInMu.Unlock()
			}()

			game.SessionStart(s)

		},
		HostSigners: []glider.Signer{hostKey}, // types match now
	}

	log.Println("Listening on port 2222...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
