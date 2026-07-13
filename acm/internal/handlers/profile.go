package handlers

import (
	"acm/internal/auth"
	"acm/internal/models"
	"acm/internal/views"
	"net/http"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetCurrentUser(r, UserRepo)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}

	transactions, err := TransactionRepo.GetTransactionsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	deposits := 0
	withdrawals := 0

	for _, t := range transactions {
		if t.Type == "deposit" {
			deposits++
		}

		if t.Type == "withdrawal" {
			withdrawals++
		}
	}

	data := struct {
		User             *models.User
		Balance          float64
		DepositCount     int
		WithdrawalCount  int
		TransactionCount int
	}{
		User:             user,
		Balance:          balance,
		DepositCount:     deposits,
		WithdrawalCount:  withdrawals,
		TransactionCount: len(transactions),
	}

	views.RenderTemplate(w, "profile.html", data)
}
