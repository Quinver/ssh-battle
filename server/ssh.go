package server

import (
	"fmt"
	"log"

	glider "github.com/gliderlabs/ssh"
	"ssh-battle/keys"
)

func StartServer() {
	hostKey, err := keys.LoadHostKey("host_key.pem")
	if err != nil {
		log.Fatal("Failed to load host key:", err)
	}

	server := &glider.Server{
		Addr: ":2222",
		Handler: func(s glider.Session) {
			fmt.Fprintf(s, "Welcome, %s! \n", s.User())
		},
		HostSigners: []glider.Signer{hostKey}, // types match now
	}

	log.Println("Listening on port 2222...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
