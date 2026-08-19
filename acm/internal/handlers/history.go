package handlers

import (
	"acm/internal/middleware"
	"acm/internal/views"
	"net/http"
)

func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	transactions, err := TransactionRepo.GetTransactionsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load transactions", http.StatusInternalServerError)
		return
	}
	views.RenderTemplate(w, r, "history.html", transactions)
}
