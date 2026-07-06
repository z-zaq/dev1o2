package handlers

import (
	"net/http"

	"acm/internal/auth"
	"acm/internal/views"
)

func DepositHandler(w http.ResponseWriter, r *http.Request) {
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
	views.RenderTemplate(w, "deposit.html", nil)
}
