package config

import (
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	ListenAddr string

	JWTSecret []byte

	SessionDuration time.Duration

	Users map[string]string

	Shell string

	CookieName string

	CookieSecure bool
}

func Load() *Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	users := parseUsers(os.Getenv("USERS"))
	if len(users) == 0 {
		log.Fatal("USERS environment variable is required (format: user1:hash1,user2:hash2)")
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	sessionDuration := 8 * time.Hour
	if d := os.Getenv("SESSION_DURATION"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			sessionDuration = parsed
		}
	}

	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"

	return &Config{
		ListenAddr:      os.Getenv("LISTEN_ADDR"),
		JWTSecret:       []byte(secret),
		SessionDuration: sessionDuration,
		Users:           users,
		Shell:           shell,
		CookieName:      "rwt_session",
		CookieSecure:    cookieSecure,
	}
}

func parseUsers(raw string) map[string]string {
	m := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		idx := strings.Index(pair, ":")
		if idx < 1 {
			log.Printf("config: skipping malformed USERS entry %q", pair)
			continue
		}
		username := pair[:idx]
		hash := pair[idx+1:]
		m[username] = hash
	}
	return m
}
