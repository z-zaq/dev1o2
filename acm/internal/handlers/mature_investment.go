package handlers

import (
	"acm/internal/middleware"
	"acm/internal/services"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// MatureInvestmentHandler lets a user claim an investment that has reached
// its end date. It recomputes the payout server-side from the same
// disclosed valuation formula used everywhere else — the client never
// supplies a payout amount.
func MatureInvestmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.UserFromContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	investmentID, err := strconv.Atoi(r.FormValue("investment_id"))
	if err != nil || investmentID <= 0 {
		http.Error(w, "Invalid investment", http.StatusBadRequest)
		return
	}

	investment, err := InvestmentRepo.GetInvestmentByID(investmentID)
	if err != nil {
		http.Error(w, "Investment not found", http.StatusNotFound)
		return
	}

	if investment.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if investment.Status != "active" {
		http.Error(w, "Investment is not active", http.StatusConflict)
		return
	}

	now := time.Now()

	if now.Before(investment.EndsAt) {
		http.Error(w, "Investment has not matured yet", http.StatusBadRequest)
		return
	}

	plan, err := PlanRepo.GetPlanByID(investment.PlanID)
	if err != nil {
		http.Error(w, "Failed to load investment plan", http.StatusInternalServerError)
		return
	}

	valuation, err := services.CalculateInvestmentValue(*investment, *plan, now)
	if err != nil {
		http.Error(w, "Failed to calculate investment value", http.StatusInternalServerError)
		return
	}

	err = InvestmentRepo.MatureInvestment(
		investment.ID,
		user.ID,
		investment.Principal,
		valuation.Profit,
	)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Investment is no longer available to claim", http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, "Failed to claim investment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/investments", http.StatusSeeOther)
}
