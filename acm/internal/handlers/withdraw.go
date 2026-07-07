package handlers

import (
	"acm/internal/auth"
	"acm/internal/views"
	"net/http"
)

func WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_, exists := auth.Sessions[cookie.Value]
	if !exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	views.RenderTemplate(w, "withdraw.html", nil)
}
