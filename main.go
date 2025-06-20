package main

import (
	"log"
	"ssh-battle/data"
	"ssh-battle/server"
)


func main() {
	// Init DB and check for errors
	data.InitDB()
	defer data.CloseDB()

	// Seed the words table (insert if not exists)
	data.SeedWords("data/words.txt")

	log.Println("Starting server...")
	server.StartServer()
}
