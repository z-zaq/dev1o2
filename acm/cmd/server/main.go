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
func renderTemplate(w http.ResponseWriter, file string) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/"+file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		log.Println("Email:", email)
		log.Println("Password:", password)
		w.Write([]byte("Login form submitted"))
		return
	}
	renderTemplate(w, "login.html")
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		type User struct {
			Name     string
			Email    string
			Password string
		}
		user := User{
			Name:     r.FormValue("name"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		log.Println("New User Registered:")
		log.Println("Name:", user.Name)
		log.Println("Name:", user.Email)
		log.Println("Password:", user.Password)
		w.Write([]byte("Registration Successful"))
		return
	}
	renderTemplate(w, "register.html")
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/contact", contactHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/register", registerHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
