package main

import (
	"acm/internal/models"
	"acm/internal/validators"
	"acm/internal/handlers"
	"log"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		log.Println("Email:", email)
		log.Println("Password:", password)
		w.Write([]byte("Login form submitted"))
		return
	}
	RenderTemplate(w, "login.html")
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := models.User{
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
		if !validators.ValidateEmail(user.Email) {
			http.Error(w, "invalid email address", http.StatusBadRequest)
			return
		}
		if !validators.ValidatePassword(user.Password) {
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
