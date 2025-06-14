package game

import (
	"fmt"
	"log"

	glider "github.com/gliderlabs/ssh"
)

func SessionStart(s glider.Session) {
	player := getOrCreatePlayer(s)
	fmt.Fprintf(s, "%s Joined \n", player.Name)
	log.Printf("%s Joined", player.Name)

	go func() {
		<-s.Context().Done()
		log.Printf("%s Left", player.Name)
	}()

	currentScene := sceneMain

	for currentScene != nil {
		currentScene = currentScene(s, player)
	}
}
