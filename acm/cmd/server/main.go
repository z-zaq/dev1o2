package main

import (
	"html/template"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home.html")
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "about.html")
}

// w.Write([]byte("About Page"))
func contactHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "contact.html")
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html")
}
func renderTemplate(w http.ResponseWriter, file string) {
	tmpl, err := template.ParseFiles("/templates/base.html", "/templates/"+file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/contact", contactHandler)
	mux.HandleFunc("/login", loginHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
