package handler

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/basemax/remote-web-terminal-socket/internal/auth"
	"github.com/basemax/remote-web-terminal-socket/internal/config"
	"golang.org/x/crypto/bcrypt"
)

var loginTmpl = template.Must(template.ParseFiles("web/templates/login.html"))

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	loginTmpl.Execute(w, nil) //nolint:errcheck
}

func Login(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			renderLoginError(w, "Username and password are required.")
			return
		}

		hash, ok := cfg.Users[username]
		if !ok {
			bcrypt.CompareHashAndPassword([]byte("$2a$12$invalid.hash.padding.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), []byte(password)) //nolint:errcheck
			renderLoginError(w, "Invalid username or password.")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
			renderLoginError(w, "Invalid username or password.")
			return
		}

		tokenStr, err := auth.IssueToken(cfg, username)
		if err != nil {
			log.Printf("login: failed to issue token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     cfg.CookieName,
			Value:    tokenStr,
			Path:     "/",
			Expires:  time.Now().Add(cfg.SessionDuration),
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteStrictMode,
		})

		http.Redirect(w, r, "/terminal", http.StatusSeeOther)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "rwt_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderLoginError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusUnauthorized)
	loginTmpl.Execute(w, map[string]string{"Error": msg}) //nolint:errcheck
}
