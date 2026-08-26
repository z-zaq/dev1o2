package views

import (
	"html/template"
	"net/http"
)
<<<<<<< HEAD
func RenderTemplate(w http.ResponseWriter, file string, data interface{}){
	tmpl, err := template.ParseFiles(
		"templates/base.html", "templates/"+file)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil{
=======

func RenderTemplate(w http.ResponseWriter, r *http.Request, file string, data interface{}) {
	loggedIn := false

	if cookie, err := r.Cookie("session"); err == nil {
		if _, exists := auth.GetSessionEmail(cookie.Value); exists {
			loggedIn = true
		}
	}

	funcMap := template.FuncMap{
		"LoggedIn": func() bool {
			return loggedIn
		},

		"mul": func(a, b float64) float64 {
			return a * b
		},
	}

	tmpl, err := template.New("base.html").
		Funcs(funcMap).
		ParseFiles(
			"templates/base.html",
			"templates/navigation.html",
			"templates/"+file,
		)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
>>>>>>> a92dbe8 (nav key)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}