package game

import (
	glider "github.com/gliderlabs/ssh"
)

type Player struct {
	Name      string
	Location  string
	Inventory []string
	Scores    []Score
}

func newPlayer(s glider.Session) *Player {
	return &Player{
		Name:      s.User(), // Get username from ssh session
		Location:  "spawn",
		Inventory: []string{},
	}
}
