package views

import (
	"acm/internal/auth"
	"html/template"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, r *http.Request, file string, data interface{}) {
	loggedIn := false
	if cookie, err := r.Cookie("session"); err == nil {
		if _, exists := auth.GetSessionEmail(cookie.Value); exists {
			loggedIn = true
		}
	}

	funcMap := template.FuncMap{
		"LoggedIn": func() bool { return loggedIn },
	}

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		"templates/base.html", "templates/"+file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
