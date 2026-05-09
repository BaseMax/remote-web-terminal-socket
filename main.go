// set GOOS=linux
// set GOARCH=amd64
// go build -o maxterminal
// chmod +x maxterminal
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/basemax/remote-web-terminal-socket/internal/auth"
	"github.com/basemax/remote-web-terminal-socket/internal/config"
	"github.com/basemax/remote-web-terminal-socket/internal/handler"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	mux.HandleFunc("/", handler.LoginPage)
	mux.HandleFunc("/login", handler.Login(cfg))
	mux.HandleFunc("/logout", handler.Logout)

	mux.Handle("/terminal", auth.Middleware(cfg)(http.HandlerFunc(handler.TerminalPage)))
	mux.Handle("/ws", auth.Middleware(cfg)(http.HandlerFunc(handler.WebSocket(cfg))))

	addr := cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("remote-web-terminal-socket listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
