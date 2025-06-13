package main

import (
	"log"
	"ssh-battle/game"
	"ssh-battle/server"
)

func main() {
	// Init DB and check for errors
	game.InitDB()
	defer game.CloseDB()

	// Seed the words table (insert if not exists)
	game.SeedWords("game/data/words.txt")

	log.Println("Starting server...")
	server.StartServer()
}
