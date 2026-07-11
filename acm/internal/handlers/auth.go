package handlers

import (
	"acm/internal/auth"
	"acm/internal/models"
	"acm/internal/repository"
	"acm/internal/validators"
	"acm/internal/views"

	"golang.org/x/crypto/bcrypt"

	// "log"
	"net/http"
)

var UserRepo *repository.UserRepository
var TransactionRepo *repository.TransactionRepository

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := models.User{
			Name:     r.FormValue("name"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
			IsAdmin:  false,
		}
		if user.Email == "admin@acm.com" {
			user.IsAdmin = true
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
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(user.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)
		err = UserRepo.CreateUser(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	views.RenderTemplate(w, "register.html", nil)
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := UserRepo.GetUserByEmail(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		err = bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(password),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		// if user.Password != password {
		// 	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		// 	return
		// }
		sessionID := auth.GenerateSessionID()
		auth.Sessions[sessionID] = user.Email

		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID,
			Path:  "/",
		})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

		// log.Println("Email:", email)
		// log.Println("password:", password)
		// return
	}
	views.RenderTemplate(w, "login.html", nil)
}
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")

	if err == nil {
		delete(auth.Sessions, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
