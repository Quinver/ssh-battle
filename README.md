# SSH Battle - Terminal Typing Game

🚀 **SSH Battle** is a multiplayer terminal-based typing game built in Go, designed to challenge your typing speed and accuracy over an SSH connection. Compete in single-player practice, head-to-head duos battles, or chat in the lobby while climbing the global leaderboard!

## Features

- **Single Player Mode**: Practice typing with randomly generated sentences and track your scores.
- **Duos Battle**: Race against another player in real-time to type sentences faster and more accurately.
- **Multiplayer Lobby**: Chat with other players and challenge them to duos matches.
- **Leaderboard**: View the top 10 global scores based on Typing Points (TP).
- **Personal Score List**: Review your top 5 scores, sorted by TP.
- **Terminal UI**: Retro-styled interface with ANSI colors, clear navigation, and intuitive controls.
- **Commands**: Use commands like `:q` to quit, `:help` for help, or `:lobby` to switch scenes anytime.
- **Score Metrics**: Tracks accuracy, words per minute (WPM), time, and a custom TP score.

## Tech Stack

- **Language**: Go
- **Libraries**:
  - gliderlabs/ssh for SSH server functionality
  - golang.org/x/term for terminal interactions
- **Database**: SQL-based (configurable, e.g., SQLite or PostgreSQL) for storing player data and scores
- **Concurrency**: Go channels and goroutines for real-time multiplayer interactions

## Installation

### Prerequisites

- Go 1.18 or higher
- Optional A SQL database, default is sqlite (e.g., mariadb, PostgreSQL)
- SSH client, accesing as client in Windows 10 default is in powershell, Windows 11 in the terminal app and most linux distros have it by default.

### Steps

1. **Clone the Repository**

   ```bash
   git clone https://github.com/quinver/ssh-battle.git
   cd ssh-battle
   ```

2. **Install Dependencies**

   ```bash
   go mod tidy
   ```

3. **Configure the Database**

Sqlite should be set up by default.
Do you want to use another database? Look around in in the /data/db.go.
I might make it configurable in the future.

4. **Build and Run**

   ```bash
   go build -o ssh-battle
   ./ssh-battle
   ```

   The server will start on the default SSH port (22) or a custom port if configured.

5. **Connect to the Game**

   Use an SSH client to connect:

   ```bash
   ssh player@localhost -p 2222
   ```

   Replace `player` with your desired username and `2222` with the configured port.

## Usage

- **Navigation**: Use ↑/↓ arrows or `j`/`k` to navigate menus, Enter to select.
- **Commands**: Type commands like `:q` (quit), `:help` (list commands), `:game` (single player), `:lobby` (multiplayer lobby), or `:duos` (duos battle) anytime.
- **Gameplay**:
  - **Single Player**: Type the displayed sentence as fast and accurately as possible.
  - **Duos**: Type `ready` to start a match against another player; race to finish first!
  - **Lobby**: Chat with others or challenge them to duos.
- **Scoring**: Scores are calculated based on accuracy, WPM, time, and a TP formula. View your top scores or the global leaderboard.

## How it looks

*Main Menu*

```
┌────────────────────────────────────────────────┐
│ 🚀 SSH Battle - Terminal Typing Game         │
└────────────────────────────────────────────────┘

Navigation Instructions:
────────────────────────
• Use ↑/↓ or j/k to navigate
• Press Enter to select
• Type commands (:help, :q, :game) anytime

Select an Option:
─────────────────
 ► Single Player Game
   Practice typing with randomly generated sentences
   Multiplayer Lobby
   Duos Battle
   Leaderboard
   Your Scores
   Quit
```

*Leaderboard*

```
┌────────────────────────────────────────────────┐
│ 🏆 SSH Battle Leaderboard                     │
└────────────────────────────────────────────────┘

Top 10 Players:
───────────────
 Rank │ Player         │ TP     │ Accuracy │ WPM   │ Time
──────┼────────────────┼────────┼──────────┼───────┼───────
 #1   │ speedster      │ 95.50  │  98.50%  │ 85.0  │  15s
 #2   │ typemaster     │ 90.25  │  97.20%  │ 80.5  │  16s
 ...
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Acknowledgements

- Inspired by classic terminal-based games and typing challenges.
- Thanks to gliderlabs/ssh for enabling SSH-based interactivity.

---
