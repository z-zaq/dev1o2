package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

type User struct {
	Name     string
	Email    string
	Password string
}

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
func validateEmail(email string) bool {
	return strings.Contains(email, "@") &&
		strings.Contains(email, ".")
}

func validatePassword(password string) bool {
	return len(password) >= 8
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := User{
			Name:     r.FormValue("name"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		if user.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if user.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}
		if user.Password == "" {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}
		if !validateEmail(user.Email) {
			http.Error(w, "invalid email address", http.StatusBadRequest)
			return
		}
		if !validatePassword(user.Password) {
			http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
			return
		}
		log.Println("New User Registered:")
		log.Println("Name:", user.Name)
		log.Println("Email:", user.Email)
		log.Println("Password:", user.Password)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
