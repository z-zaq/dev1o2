package handlers

import (
	"net/http"
	"strconv"

	"acm/internal/auth"
	"acm/internal/models"
	"acm/internal/views"
	// "acm/internal/repository"
)

func DepositHandler(w http.ResponseWriter, r *http.Request) {
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
	// get current user
	user, err := UserRepo.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	// handle form submission
	if r.Method == http.MethodPost {
		amountStr := r.FormValue("amount")

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}
		if amount <= 0 {
			http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
			return
		}
		transaction := models.Transaction{
			UserID: user.ID,
			Type:   "deposit",
			Amount: amount,
		}
		err = TransactionRepo.CreateTransaction(transaction)
		if err != nil {
			http.Error(w, "Failed to create deposit", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
	views.RenderTemplate(w, "deposit.html", nil)
}
