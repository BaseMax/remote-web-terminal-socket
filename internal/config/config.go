package config

import (
	"log"
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration.
type Config struct {
	// HTTP listen address, e.g. ":8080"
	ListenAddr string

	// JWT signing secret — MUST be set via env var JWT_SECRET
	JWTSecret []byte

	// Session duration for the JWT cookie
	SessionDuration time.Duration

	// Allowed credentials: map[username]bcrypt-hashed-password
	// Populated from env USERNAME / PASSWORD_HASH pairs.
	// Format: "user1:hash1,user2:hash2"
	Users map[string]string

	// Shell to spawn in the PTY (default: /bin/bash)
	Shell string

	// Name of the cookie used to carry the JWT
	CookieName string

	// Secure flag on the cookie (set true behind HTTPS / nginx)
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

// parseUsers parses "user1:hash1,user2:hash2" into a map.
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
