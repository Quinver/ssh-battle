package server

import (
	"log"
	"ssh-battle/keys"
	"ssh-battle/player"
	"ssh-battle/scenes"
	"strings"
	"sync"

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
			return player.CheckPassword(ctx.User(), password)
		},
		Handler: func(s glider.Session) {
			username := s.User()

			loggedInMu.Lock()
			if strings.ToLower((s.User())) == "root" {
				loggedInMu.Unlock()
				s.Write([]byte("Can't login as root to avoid bots from scanning this session. Try running something like \"ssh Username@quinver.dev -p 2222\"...\n"))
				s.Close()
				return
			}
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

			scenes.SessionStart(s)
		},
		HostSigners: []glider.Signer{hostKey}, // types match now
	}

	log.Println("Listening on port 2222...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
