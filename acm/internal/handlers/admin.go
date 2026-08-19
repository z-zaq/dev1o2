package handlers

import (
	"acm/internal/views"
	"net/http"
	"strconv"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}

		userID, err := strconv.Atoi(r.FormValue("user_id"))
		if err != nil || userID <= 0 {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		role := r.FormValue("role")
		if role != "user" && role != "admin" {
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}

		user, err := UserRepo.GetUserByID(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if user.Role == "admin" && role == "user" {
			adminCount, err := UserRepo.CountAdmins()
			if err != nil {
				http.Error(w, "Failed to check admin count", http.StatusInternalServerError)
				return
			}

			if adminCount <= 1 {
				http.Error(w, "Cannot demote the last administrator", http.StatusBadRequest)
				return
			}
		}

		if err := UserRepo.UpdateUserRole(userID, role); err != nil {
			http.Error(w, "Failed to update user role", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	users, err := UserRepo.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	transactions, err := TransactionRepo.GetAllTransactions()
	if err != nil {
		http.Error(w, "Failed to load transactions", http.StatusInternalServerError)
		return
	}

	data := struct {
		Users        interface{}
		Transactions interface{}
	}{
		Users:        users,
		Transactions: transactions,
	}

	views.RenderTemplate(w, r, "admin.html", data)
}
