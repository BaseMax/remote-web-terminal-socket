package main

import (
	"log"
	"net/http"
	"os"

	"github.com/basemax/remote-web-terminal/internal/auth"
	"github.com/basemax/remote-web-terminal/internal/config"
	"github.com/basemax/remote-web-terminal/internal/handler"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// Static assets
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Public routes
	mux.HandleFunc("/", handler.LoginPage)
	mux.HandleFunc("/login", handler.Login(cfg))
	mux.HandleFunc("/logout", handler.Logout)

	// Protected routes — require valid JWT cookie
	mux.Handle("/terminal", auth.Middleware(cfg)(http.HandlerFunc(handler.TerminalPage)))
	mux.Handle("/ws", auth.Middleware(cfg)(http.HandlerFunc(handler.WebSocket(cfg))))

	addr := cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("remote-web-terminal listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
