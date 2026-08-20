package handlers

import (
	"acm/internal/middleware"
	"acm/internal/models"
	"acm/internal/services"
	"acm/internal/views"
	"net/http"
	"time"
)

type InvestmentView struct {
	Investment models.Investment
	Plan       *models.Plan
	Valuation  services.InvestmentValuation
}

func InvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r)

	investments, err := InvestmentRepo.GetInvestmentsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load investments", http.StatusInternalServerError)
		return
	}

	viewsData := make([]InvestmentView, 0, len(investments))

	for _, investment := range investments {
		plan, err := PlanRepo.GetPlanByID(investment.PlanID)
		if err != nil {
			http.Error(w, "Failed to load investment plan", http.StatusInternalServerError)
			return
		}

		valuation, err := services.CalculateInvestmentValue(
			investment,
			*plan,
			time.Now(),
		)
		if err != nil {
			http.Error(w, "Failed to calculate investment value", http.StatusInternalServerError)
			return
		}

		viewsData = append(viewsData, InvestmentView{
			Investment: investment,
			Plan:       plan,
			Valuation:  valuation,
		})
	}

	views.RenderTemplate(w, r, "investments.html", viewsData)
}
