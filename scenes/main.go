package scenes

import (
	"fmt"
	"log"
	"ssh-battle/player"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

func Main(s glider.Session, p *player.Player) Scene {
	shell := term.NewTerminal(s, "> ")
	clearTerminal(shell)

	shell.Write([]byte("You're in the main scene. Type :help to find more scenes\n"))

	for {
		_, nextScene, done := SafeReadInput(shell, s, p)
		if done {
			return nextScene
		}
	}
}

func SessionStart(s glider.Session) {
	player := player.GetOrCreatePlayer(s)
	fmt.Fprintf(s, "%s Joined \n", player.Name)
	log.Printf("%s Joined", player.Name)

	go func() {
		<-s.Context().Done()
		log.Printf("%s Left", player.Name)
	}()

	currentScene := Main

	for currentScene != nil {
		currentScene = currentScene(s, player)
	}
}
