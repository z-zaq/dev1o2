package handlers

import (
	"acm/internal/models"
	"acm/internal/services"
	"acm/internal/views"
	"net/http"
	"strconv"
	"time"
)

type AdminInvestmentView struct {
	Investment models.Investment
	Plan       *models.Plan
	Valuation  services.InvestmentValuation
	OwnerName  string
	OwnerEmail string
}

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

	investments, err := InvestmentRepo.GetAllInvestments()
	if err != nil {
		http.Error(w, "Failed to load investments", http.StatusInternalServerError)
		return
	}

	usersByID := make(map[int]models.User, len(users))
	for _, u := range users {
		usersByID[u.ID] = u
	}

	investmentViews := make([]AdminInvestmentView, 0, len(investments))
	for _, investment := range investments {
		plan, err := PlanRepo.GetPlanByID(investment.PlanID)
		if err != nil {
			http.Error(w, "Failed to load investment plan", http.StatusInternalServerError)
			return
		}

		valuation, err := services.CalculateInvestmentValue(investment, *plan, time.Now())
		if err != nil {
			http.Error(w, "Failed to calculate investment value", http.StatusInternalServerError)
			return
		}

		owner := usersByID[investment.UserID]

		investmentViews = append(investmentViews, AdminInvestmentView{
			Investment: investment,
			Plan:       plan,
			Valuation:  valuation,
			OwnerName:  owner.Name,
			OwnerEmail: owner.Email,
		})
	}

	data := struct {
		Users        interface{}
		Transactions interface{}
		Investments  interface{}
	}{
		Users:        users,
		Transactions: transactions,
		Investments:  investmentViews,
	}

	views.RenderTemplate(w, r, "admin.html", data)
}
