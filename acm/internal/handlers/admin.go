package handlers

import (
	"acm/internal/views"
	"net/http"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
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
	views.RenderTemplate(w, "admin.html", data)
}
