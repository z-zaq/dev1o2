package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/contact", contactHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/about.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tmpl.Execute(w, nil)
}

// w.Write([]byte("About Page"))
func contactHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/contact.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
	// w.Write([]byte("Contact Page"))
}
