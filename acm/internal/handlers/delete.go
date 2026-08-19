package handlers

import (
	"acm/internal/auth"
	"acm/internal/middleware"
	"acm/internal/views"
	"net/http"
)

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	if r.Method == http.MethodPost {
		err := UserRepo.DeleteUser(user.ID)
		if err != nil {
			http.Error(w, "Failed to delete account", http.StatusInternalServerError)
			return
		}

		if cookie, err := r.Cookie("session"); err == nil {
			auth.DeleteSession(cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	views.RenderTemplate(w, r, "delete_account.html", user)
}
