package handlers

import (
	"acm/internal/middleware"
	"acm/internal/models"
	"acm/internal/views"
	"net/http"
	"strconv"
	"time"
)

func WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

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
		balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
		if err != nil {
			http.Error(w, "Failed to calculate balance", http.StatusInternalServerError)
			return
		}
		if amount > balance {
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
		transaction := models.Transaction{
			UserID:      user.ID,
			Type:        "withdrawal",
			Amount:      amount,
			Description: "Withdrawal",
			CreatedAt:   time.Now(),
		}
		err = TransactionRepo.CreateTransaction(transaction)
		if err != nil {
			http.Error(w, "Failed to create withdrawal", http.StatusInternalServerError)
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
	views.RenderTemplate(w, r, "withdraw.html", data)
}
