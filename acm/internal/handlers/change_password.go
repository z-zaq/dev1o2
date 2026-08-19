package handlers

import (
	"acm/internal/middleware"
	"acm/internal/views"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	if r.Method == http.MethodPost {
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		if newPassword != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(currentPassword),
		)
		if err != nil {
			http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(newPassword),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		err = UserRepo.UpdatePassword(
			user.ID,
			string(hashedPassword),
		)
		if err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	views.RenderTemplate(w, r, "change_password.html", nil)
}
