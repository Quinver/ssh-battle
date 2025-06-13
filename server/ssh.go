package server

import (
	"log"

	glider "github.com/gliderlabs/ssh"
	"ssh-battle/keys"
	"ssh-battle/game"
)

func StartServer() {
	hostKey, err := keys.LoadHostKey("host_key.pem")
	if err != nil {
		log.Fatal("Failed to load host key:", err)
	}

	server := &glider.Server{
		Addr: ":2222",
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
