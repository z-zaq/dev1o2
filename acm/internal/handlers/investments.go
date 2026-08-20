package handlers

import (
	"acm/internal/middleware"
	"acm/internal/models"
	"acm/internal/views"
	"net/http"
)

type InvestmentView struct {
	Investment models.Investment
	Plan       *models.Plan
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

		viewsData = append(viewsData, InvestmentView{
			Investment: investment,
			Plan:       plan,
		})
	}

	views.RenderTemplate(w, r, "investments.html", viewsData)
}
