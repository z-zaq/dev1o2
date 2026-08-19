package handlers

import (
	"acm/internal/middleware"
	"acm/internal/models"
	"acm/internal/views"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load balance", http.StatusInternalServerError)
		return
	}
	transactions, err := TransactionRepo.GetTransactionsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load transaction history", http.StatusInternalServerError)
		return
	}
	recent := transactions
	if len(recent) > 5 {
		recent = recent[:5]
	}
	data := struct {
		User    *models.User
		Balance float64
		Recent  []models.Transaction
	}{
		User:    user,
		Balance: balance,
		Recent:  recent,
	}
	views.RenderTemplate(w, r, "dashboard.html", data)
}
