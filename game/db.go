package game

import (
	"bufio"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() {

	err := os.MkdirAll("game/data", 0755)
	// Make sure the directory exists
	if err != nil {
		log.Fatal("failed to create data directory:", err)
	}

	var err2 error
	db, err2 = sql.Open("sqlite3", "game/data/game.db")
	if err2 != nil {
		log.Fatal(err2)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT
	);

	CREATE TABLE IF NOT EXISTS scores (
	  	id INTEGER PRIMARY KEY AUTOINCREMENT,
	  	player_id INTEGER NOT NULL,
	  	accuracy REAL,
	  	wpm REAL,
		tp REAL,
	  	duration INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, 
		FOREIGN KEY(player_id) REFERENCES players(id) ON DELETE CASCADE
	);


`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}

func CloseDB() {
	db.Close()
}

func SeedWords(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO words (word) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for scanner.Scan() {
		word := scanner.Text()
		if _, err := stmt.Exec(word); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetWordsFromDB() ([]string, error) {
	rows, err := db.Query("SELECT word FROM words")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbWords []string
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			return nil, err
		}
		dbWords = append(dbWords, w)
	}
	return dbWords, nil
}
