package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/basemax/remote-web-terminal/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsKey contextKey = "claims"

// Claims is the JWT payload.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// IssueToken creates a signed JWT for the given user.
func IssueToken(cfg *config.Config, username string) (string, error) {
	now := time.Now()
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.SessionDuration)),
			Issuer:    "remote-web-terminal",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.JWTSecret)
}

// ParseToken validates a JWT string and returns the claims.
func ParseToken(cfg *config.Config, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return cfg.JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// Middleware returns an http.Handler that enforces JWT authentication via cookie.
// On failure it redirects to the login page.
func Middleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			claims, err := ParseToken(cfg, cookie.Value)
			if err != nil {
				// Clear the invalid cookie
				http.SetCookie(w, expiredCookie(cfg.CookieName, cfg.CookieSecure))
				redirectToLogin(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext retrieves the claims from a request context.
func FromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// For WebSocket upgrade requests return 401 instead of redirect
	if r.Header.Get("Upgrade") == "websocket" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func expiredCookie(name string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
}
