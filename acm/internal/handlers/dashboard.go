package handlers

import (
	"acm/internal/auth"
	"acm/internal/models"
	"acm/internal/views"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	email, exists := auth.Sessions[cookie.Value]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := UserRepo.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load balance", http.StatusInternalServerError)
		return
	}
	data := struct {
		User    *models.User
		Balance float64
	}{
		User:    user,
		Balance: balance,
	}
	views.RenderTemplate(w, "dashboard.html", data)
}
