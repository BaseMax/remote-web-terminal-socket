package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/basemax/remote-web-terminal/internal/config"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		return r.Header.Get("Origin") == "https://"+r.Host ||
			r.Header.Get("Origin") == "http://"+r.Host
	},
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type resizeMsg struct {
	Type string `json:"type"` // "resize"
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

func WebSocket(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws: upgrade error: %v", err)
			return
		}
		defer conn.Close()

		shell := cfg.Shell
		if shell == "" {
			shell = "/bin/bash"
		}

		cmd := exec.Command(shell)
		cmd.Env = append(os.Environ(), "TERM=xterm-256color")

		ptmx, err := pty.Start(cmd)
		if err != nil {
			log.Printf("ws: pty start error: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte("\r\nFailed to start shell: "+err.Error()+"\r\n")) //nolint:errcheck
			return
		}
		defer func() {
			ptmx.Close()
			cmd.Process.Kill() //nolint:errcheck
			cmd.Wait()         //nolint:errcheck
		}()

		var writeMu sync.Mutex

		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := ptmx.Read(buf)
				if n > 0 {
					writeMu.Lock()
					werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n])
					writeMu.Unlock()
					if werr != nil {
						return
					}
				}
				if err != nil {
					return
				}
			}
		}()

		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			switch msgType {
			case websocket.BinaryMessage:
				ptmx.Write(data) //nolint:errcheck

			case websocket.TextMessage:
				var msg resizeMsg
				if jsonErr := json.Unmarshal(data, &msg); jsonErr == nil && msg.Type == "resize" {
					pty.Setsize(ptmx, &pty.Winsize{
						Cols: msg.Cols,
						Rows: msg.Rows,
					}) //nolint:errcheck
				} else {
					ptmx.Write(data) //nolint:errcheck
				}
			}
		}
	}
}
