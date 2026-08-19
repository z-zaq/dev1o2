package handlers

import (
	"acm/internal/middleware"
	"acm/internal/views"
	"net/http"
	"strconv"
)

func TransferHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	if r.Method == http.MethodPost {
		recipientEmail := r.FormValue("recipient")
		amountStr := r.FormValue("amount")

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}
		if recipientEmail == user.Email {
			http.Error(w, "Cannot transfer to yourself", http.StatusBadRequest)
			return
		}
		recipient, err := UserRepo.GetUserByEmail(recipientEmail)
		if err != nil {
			http.Error(w, "Recipient not found", http.StatusBadRequest)
			return
		}
		balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
		if err != nil {
			http.Error(w, "Failed to get balance", http.StatusInternalServerError)
			return
		}
		if amount > balance {
			http.Error(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
		err = TransactionRepo.Transfer(
			user.ID,
			recipient.ID,
			amount,
			user.Email,
			recipientEmail,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	views.RenderTemplate(w, r, "transfer.html", data)
}
