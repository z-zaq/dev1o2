package handlers

import (
	"net/http"
	"strconv"
	"time"

	"acm/internal/auth"
	"acm/internal/models"
	"acm/internal/views"
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
	user, err := UserRepo.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
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
			UserID:      user.ID,
			Type:        "deposit",
			Amount:      amount,
			Description: "Deposit",
			CreatedAt:   time.Now(),
		}
		err = TransactionRepo.CreateTransaction(transaction)
		if err != nil {
			http.Error(w, "Failed to create deposit", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load balance", http.StatusInternalServerError)
		return
	}
	data := struct {
		Balance float64
	}{
		Balance: balance,
	}
	views.RenderTemplate(w, "deposit.html", data)
}
