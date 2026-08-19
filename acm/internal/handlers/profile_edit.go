package handlers

import (
	"acm/internal/middleware"
	"acm/internal/views"
	"net/http"
)

func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

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

		err := UserRepo.UpdateUser(*user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	views.RenderTemplate(w, r, "edit_profile.html", user)
}
