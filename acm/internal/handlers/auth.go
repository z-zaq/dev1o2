package handlers

import (
	"acm/internal/models"
	"acm/internal/validators"
	"acm/internal/views"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		log.Println("Email:", email)
		log.Println("password:", password)
		return
	}
	views.RenderTemplate(w, "login.html")
}
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := models.User{
			Name:    r.FormValue("name"),
			Email:   r.FormValue("email"),
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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	views.RenderTemplate(w, "register.html")
}
