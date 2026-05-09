package handler

import (
	"html/template"
	"net/http"

	"github.com/basemax/remote-web-terminal/internal/auth"
)

var terminalTmpl = template.Must(template.ParseFiles("web/templates/terminal.html"))

func TerminalPage(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	username := ""
	if claims != nil {
		username = claims.Username
	}
	terminalTmpl.Execute(w, map[string]string{"Username": username}) //nolint:errcheck
}
