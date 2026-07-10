package handlers

import (
	"acm/internal/auth"
	"acm/internal/views"
	"net/http"
)

func EditProfileHandler(w http.ResponseWriter, r *http.Request) {

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
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		user.Name = name
		user.Email = email

		err = UserRepo.UpdateUser(*user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// if err != nil {
		// 	http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		// 	return
		// }
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	views.RenderTemplate(w, "edit_profile.html", user)
}
