package game

import (
	"log"
	"strings"
	"sync"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	mu        sync.Mutex
	players   = make(map[string]*Player)
	passwords = make(map[string]string)
)

type Player struct {
	Name   string
	Scores []Score
}

func CheckPassword(username, password string) bool {
	mu.Lock()
	defer mu.Unlock()

	stored, ok := passwords[username]
	if !ok {
		// New user accept any password for now, setting password after login
		return true
	}
	return CheckPasswordHash(password, stored)
}

func CreateNewPassword(s glider.Session) {
	mu.Lock()
	_, exists := passwords[s.User()]
	mu.Unlock()

	if exists {
		return
	}

	shell := term.NewTerminal(s, "> ")

	for {
		shell.Write([]byte("You don't have a password, create one:\n"))
		pass, err := shell.ReadLine()
		if err != nil {
			return
		}
		pass = strings.TrimSpace(pass)

		shell.Write([]byte("Confirm password:\n"))
		confPass, err := shell.ReadLine()
		if err != nil {
			return
		}
		confPass = strings.TrimSpace(confPass)

		if pass != confPass {
			shell.Write([]byte("Passwords do not match. Try again.\n\n"))
			continue
		}

		hash, err := HashPassword(pass)
		if err != nil {
			shell.Write([]byte("Internal error. Try again.\n"))
			continue
		}

		mu.Lock()
		passwords[s.User()] = hash
		mu.Unlock()

		shell.Write([]byte("Password set! You are now logged in.\n"))
		break
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func getOrCreatePlayer(s glider.Session) *Player {
	key := s.PublicKey()
	if key == nil {
		log.Println("No PublicKey found for", s.User())
		mu.Lock()
		defer mu.Unlock()
		name := s.User()
		player, exists := players[name]
		if !exists {
			player = &Player{
				Name:   name,
				Scores: []Score{},
			}
			players[name] = player
		}
		return player
	}
	fingerprint := ssh.FingerprintSHA256(key)

	mu.Lock()
	defer mu.Unlock()

	player, exists := players[fingerprint]
	if !exists {
		player = &Player{
			Name:   s.User(),
			Scores: []Score{},
		}
		players[fingerprint] = player
	}

	return player
}
