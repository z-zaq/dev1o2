package handlers

import (
	"acm/internal/middleware"
	"acm/internal/models"
	"acm/internal/views"
	"net/http"
	"strconv"
	"time"
)

func InvestHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}

		planID, err := strconv.Atoi(r.FormValue("plan_id"))
		if err != nil || planID <= 0 {
			http.Error(w, "Invalid investment plan", http.StatusBadRequest)
			return
		}

		amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
		if err != nil || amount <= 0 {
			http.Error(w, "Invalid investment amount", http.StatusBadRequest)
			return
		}

		plan, err := PlanRepo.GetPlanByID(planID)
		if err != nil {
			http.Error(w, "Investment plan not found", http.StatusNotFound)
			return
		}

		if amount < plan.MinDeposit {
			http.Error(w, "Investment amount is below the minimum deposit", http.StatusBadRequest)
			return
		}

		if amount > plan.MaxDeposit {
			http.Error(w, "Investment amount exceeds the maximum deposit", http.StatusBadRequest)
			return
		}

		balance, err := TransactionRepo.GetBalanceByUserID(user.ID)
		if err != nil {
			http.Error(w, "Failed to check account balance", http.StatusInternalServerError)
			return
		}

		if amount > balance {
			http.Error(w, "Insufficient available balance", http.StatusBadRequest)
			return
		}

		startedAt := time.Now()
		endsAt := startedAt.AddDate(0, 0, plan.Duration)

		investment := models.Investment{
			UserID:    user.ID,
			PlanID:    plan.ID,
			Principal: amount,
			StartedAt: startedAt,
			EndsAt:    endsAt,
			Status:    "active",
		}

		if err := InvestmentRepo.CreateInvestmentWithFunding(investment); err != nil {
			http.Error(w, "Failed to create investment", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/investments", http.StatusSeeOther)
		return
	}

	plans, err := PlanRepo.GetAllPlans()
	if err != nil {
		http.Error(w, "Failed to load investment plans", http.StatusInternalServerError)
		return
	}

	views.RenderTemplate(w, r, "invest.html", plans)
}
