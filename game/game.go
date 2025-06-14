package game

import (
	"fmt"
	"log"

	glider "github.com/gliderlabs/ssh"
)

func SessionStart(s glider.Session) {
	player := newPlayer(s)
	fmt.Fprintf(s, "%s Joined \n", player.Name)

	log.Printf("%s Joined", player.Name)

	currentScene := sceneMain

	for currentScene != nil {
		currentScene = currentScene(s, player)
	}
}
