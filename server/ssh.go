package server

import (
	"io"
	"log"
	"os"
	"ssh-battle/keys"
	"ssh-battle/player"
	"ssh-battle/scenes"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	glider "github.com/gliderlabs/ssh"
)

var loggedInUsers = make(map[string]bool)
var loggedInMu sync.Mutex

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func StartServer() {
	hostKey, err := keys.LoadHostKey("host_key.pem")
	if err != nil {
		log.Fatal("Failed to load host key:", err)
	}

	server := &glider.Server{
		Addr: ":2222",
		PasswordHandler: func(ctx glider.Context, password string) bool {
			return player.CheckPassword(ctx.User(), password)
		},
		Handler: func(s glider.Session) {
			// Use pty
			ptyReq, winCh, isPty := s.Pty()

			if !isPty {
				s.Write([]byte("This application requires a PTY.\n"))
				s.Exit(1)
				return
			}

			f, tty, err := pty.Open()
			if err != nil {
				s.Write([]byte("Could not create PTY\n"))
				s.Exit(1)
				return
			}
			defer tty.Close()

			setWinsize(f, ptyReq.Window.Width, ptyReq.Window.Height)
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()

			// Pipe session I/O through pty
			go io.Copy(f, s)
			go io.Copy(s, f)

			username := strings.ToLower(s.User())

			loggedInMu.Lock()
			if strings.ToLower((s.User())) == "root" {
				loggedInMu.Unlock()
				s.Write([]byte("Can't login as root to avoid bots from scanning this session. Try running something like \"ssh Username@quinver.dev -p 2222\"...\n"))
				s.Close()
				return
			}
			if loggedInUsers[username] {
				loggedInMu.Unlock()
				s.Write([]byte("User already logged in elsewhere. Disconnecting...\n"))
				s.Close()
				return
			}

			loggedInUsers[username] = true
			loggedInMu.Unlock()

			// Delete user from currently save users after session ends
			defer func() {
				loggedInMu.Lock()
				delete(loggedInUsers, username)
				loggedInMu.Unlock()
			}()

			scenes.SessionStart(s)
		},
		HostSigners: []glider.Signer{hostKey}, // types match now
	}

	log.Println("Listening on port 2222...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
