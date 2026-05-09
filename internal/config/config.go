package config

import (
	"bufio"
	"log"
	"os"
	"runtime"
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

// loadDotEnv reads a .env file from the working directory and sets any
// variables that are not already present in the process environment.
// It is safe to call even when the file does not exist.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // file absent — not an error
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// Strip surrounding single or double quotes
		if len(val) >= 2 {
			if (val[0] == '\'' && val[len(val)-1] == '\'') ||
				(val[0] == '"' && val[len(val)-1] == '"') {
				val = val[1 : len(val)-1]
			}
		}

		// Only set if not already in environment (real env vars take priority)
		if os.Getenv(key) == "" {
			os.Setenv(key, val) //nolint:errcheck
		}
	}
}

func Load() *Config {
	loadDotEnv(".env")
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
		if runtime.GOOS == "windows" {
			shell = "powershell.exe"
		} else {
			shell = "/bin/bash"
		}
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
