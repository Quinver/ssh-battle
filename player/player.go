package player

import (
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	glider "github.com/gliderlabs/ssh"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"ssh-battle/data"
)

var mu sync.Mutex

type Player struct {
	ID       int
	Name     string
	Scores   []Score
	Session  glider.Session
	Messages chan string
	Ready    bool

	Shell  *term.Terminal       
	WinCh  <-chan glider.Window 	
	PtyReq *glider.Pty          
}

func playerExists(username string) (bool, error) {
	var exists bool
	err := data.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM players WHERE username = ?)", username).Scan(&exists)
	return exists, err
}

func CheckPassword(username, password string) bool {
	var hash sql.NullString
	err := data.DB.QueryRow("SELECT password_hash FROM players WHERE username = ?", username).Scan(&hash)
	if err == sql.ErrNoRows {
		// User not found, allow login
		return true
	}
	if err != nil {
		log.Println("CheckPassword DB error:", err)
		return false
	}
	// If password hash is NULL, allow login (no password set)
	if !hash.Valid {
		return true
	}
	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(password))
	return err == nil
}

func CreateNewPassword(s glider.Session) {
	shell := term.NewTerminal(s, "> ")

	for {
		shell.Write([]byte("You don't have a password, create one:\n"))
		pass, err := shell.ReadPassword("Password(Or enter when new user): ")
		if err != nil {
			return
		}
		pass = strings.TrimSpace(pass)

		shell.Write([]byte("Confirm password:\n"))
		confPass, err := shell.ReadPassword("Password(Or enter when new user): ")
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

		_, err = data.DB.Exec("INSERT OR REPLACE INTO players (username, password_hash) VALUES (?, ?)", s.User(), hash)
		if err != nil {
			log.Println("DB insert error:", err)
			shell.Write([]byte("Failed to save password. Try again later.\n"))
			return
		}

		log.Printf("%s Has set a new password", s.User())
		shell.Write([]byte("Password set! You are now logged in.\n"))
		break
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func getPasswordHash(username string) (sql.NullString, error) {
	var hash sql.NullString
	err := data.DB.QueryRow("SELECT password_hash FROM players WHERE username = ?", username).Scan(&hash)
	return hash, err
}

func GetOrCreatePlayer(s glider.Session) *Player {
	mu.Lock()
	defer mu.Unlock()

	name := s.User()

	exists, err := playerExists(name)
	if err != nil {
		log.Println("DB error checking player existence:", err)
		return nil
	}

	if !exists {
		_, err := data.DB.Exec("INSERT INTO players (username) VALUES (?)", name)
		if err != nil {
			log.Println("DB error inserting new player:", err)
			return nil
		}
	}

	// Check if password_hash is NULL (no password set)
	passHash, err := getPasswordHash(name)
	if err != nil {
		log.Println("DB error getting password hash:", err)
		return nil
	}

	if !passHash.Valid {
		// password_hash is NULL → force password creation before login
		CreateNewPassword(s)
	}

	// Retrieve player id and username
	var id int
	err = data.DB.QueryRow("SELECT id, username FROM players WHERE username = ?", name).Scan(&id, &name)
	if err != nil {
		log.Println("DB error retrieving player:", err)
		return nil
	}

	player := &Player{
		ID:   id,
		Name: name,
	}

	player.Messages = make(chan string, 10)

	// Assign player scores
	scores, err := getScoresForPlayer(player.ID)
	if err != nil {
		log.Println("DB error retrieving scores:", err)
	} else {
		player.Scores = scores
	}

	return player
}

func getScoresForPlayer(playerID int) ([]Score, error) {
	rows, err := data.DB.Query("SELECT id, accuracy, wpm, tp, duration, created_at FROM scores WHERE player_id = ?", playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var s Score
		var createdAt time.Time
		err := rows.Scan(&s.ID, &s.Accuracy, &s.WPM, &s.TP, &s.Duration, &createdAt)
		if err != nil {
			return nil, err
		}
		s.CreatedAt = &createdAt
		scores = append(scores, s)
	}
	return scores, nil
}

func (p *Player) SendMessage(msg string) {
	if p == nil {
		return
	}
	select {
	case p.Messages <- msg:
	default:
		log.Printf("Dropping message for %s (channel full)", p.Name)
	}
}
