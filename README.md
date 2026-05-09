# Remote Web Terminal

A **production-grade**, browser-based terminal over WebSocket built in Go.

- Full interactive PTY (pseudo-terminal) — works with `vim`, `htop`, `tmux`, etc.
- JWT authentication via HttpOnly cookie (no credentials in URL or JS)
- Login page with bcrypt password verification (constant-time, timing-attack resistant)
- xterm.js v5 frontend with 256-colour support, copy/paste, search, web links
- Designed to run behind **nginx** (WebSocket proxy config included)
- Docker & docker-compose ready

---

## Quick Start

### 1. Generate a password hash

```bash
make hash PASS=mysecretpassword
# → $2a$10$...
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env — set JWT_SECRET and USERS
```

`.env` example:
```
JWT_SECRET=<openssl rand -hex 32>
USERS=admin:$2a$10$<your-hash>
COOKIE_SECURE=false      # set true behind HTTPS
```

> Multiple users: `USERS=alice:<hash1>,bob:<hash2>`

### 3. Run

```bash
# Native (Linux/macOS)
source .env && make run

# Docker
docker-compose up --build
```

Open `http://localhost:8080` → log in → terminal.

---

## Environment Variables

| Variable           | Required | Default      | Description                                         |
|--------------------|----------|--------------|-----------------------------------------------------|
| `JWT_SECRET`       | **yes**  | —            | HMAC-SHA256 signing key. Use a long random string.  |
| `USERS`            | **yes**  | —            | Comma-separated `user:bcrypt-hash` pairs.           |
| `LISTEN_ADDR`      | no       | `:8080`      | `host:port` to bind.                                |
| `SHELL`            | no       | `/bin/bash`  | Shell binary launched in the PTY.                   |
| `SESSION_DURATION` | no       | `8h`         | Go duration string for JWT/cookie lifetime.         |
| `COOKIE_SECURE`    | no       | `false`      | Set `true` when served over HTTPS.                  |

---

## Nginx

Copy `nginx.conf.example` to your nginx config directory. The `/ws` location handles WebSocket upgrade with a 1-hour keep-alive.

---

## Security notes

| Concern | Mitigation |
|---|---|
| Credentials in transit | HttpOnly + Secure cookie; JWT signed with HS256 |
| Brute-force login | bcrypt (cost 12); add nginx rate-limiting |
| Timing attacks | Constant-time bcrypt dummy on unknown username |
| CSRF | SameSite=Strict cookie; POST-only login |
| WebSocket origin | Origin checked against Host header |
| XSS | `Content-Security-Policy` header in nginx example |

---

## Keyboard shortcuts

| Shortcut | Action |
|---|---|
| `Ctrl+Shift+C` | Copy selection |
| `Ctrl+Shift+V` | Paste from clipboard |
| `Ctrl+Shift+A` | Select all |
| `Ctrl+Shift+F` | Toggle search bar |
| Right-click | Context menu (copy / paste / select-all / clear) |

---

## Project layout

```
.
├── main.go
├── internal/
│   ├── auth/auth.go               # JWT issue / verify / middleware
│   ├── config/config.go           # Environment-based config
│   └── handler/
│       ├── login.go               # Login / logout handlers
│       ├── terminal.go            # Terminal page handler
│       └── websocket.go           # WebSocket ↔ PTY bridge
├── web/
│   ├── templates/login.html
│   ├── templates/terminal.html
│   └── static/css/ js/
├── tools/hashpw/main.go           # bcrypt hash helper
├── Dockerfile
├── docker-compose.yml
├── nginx.conf.example
└── Makefile
```

License MIT

Copyright 2026, Seyyed Ali Mohammadiyeh (MAX BASE)
