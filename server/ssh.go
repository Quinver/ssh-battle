package server

import (
	"log"

	"ssh-battle/game"
	"ssh-battle/keys"

	glider "github.com/gliderlabs/ssh"
)

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
			game.SessionStart(s)
		},
		HostSigners: []glider.Signer{hostKey}, // types match now
	}

	log.Println("Listening on port 2222...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
