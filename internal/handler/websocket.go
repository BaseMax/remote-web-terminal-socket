package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/basemax/remote-web-terminal/internal/config"
	"github.com/basemax/remote-web-terminal/internal/ptybridge"
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

		bridge, err := ptybridge.Start(cfg.Shell)
		if err != nil {
			log.Printf("ws: pty start error: %v", err)
			msg := "\r\n\x1b[31m[ERROR] Failed to start shell (" + cfg.Shell + "): " + err.Error() + "\x1b[0m\r\n"
			conn.WriteMessage(websocket.BinaryMessage, []byte(msg)) //nolint:errcheck
			// Hold connection open so xterm can render the error before the
			// client's reconnect logic fires.
			time.Sleep(3 * time.Second)
			return
		}
		defer bridge.Close()

		var writeMu sync.Mutex

		// PTY → WebSocket (binary frames)
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := bridge.Read(buf)
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

		// WebSocket → PTY
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			switch msgType {
			case websocket.BinaryMessage:
				bridge.Write(data) //nolint:errcheck

			case websocket.TextMessage:
				var msg resizeMsg
				if jsonErr := json.Unmarshal(data, &msg); jsonErr == nil && msg.Type == "resize" {
					bridge.Resize(msg.Cols, msg.Rows) //nolint:errcheck
				} else {
					bridge.Write(data) //nolint:errcheck
				}
			}
		}
	}
}
