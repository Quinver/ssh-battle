# SSH Battle

A fun terminal-based typing game playable over SSH.

## Features

- SSH server where players can join and play
- Random sentence generation using a word list
- SQLite database used to store and load words
- Easily expandable word source (supports external word files)
- Built with Go

## Dependencies
- go
## Setup

1. Clone the repo:
    ```bash
    git clone https://github.com/quinver/ssh-battle.git
    cd ssh-battle
    ```

2. Build and run:
    ```bash
    go run main.go
    ```
    If you get errors try:
    ```bash
    go mod tidy
    ```

4. Connect via SSH:
    ```bash
    ssh youruser@localhost -p 2222
    ```

## Word Seeding

- Words are stored in a SQLite database at `game/data/game.db`
- You can seed words from a plain text file (one word per line):
    ```go
    game.SeedWords("path/to/wordlist.txt")
    ```
