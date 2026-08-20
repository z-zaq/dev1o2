package handlers

import (
	"acm/internal/views"
	"net/http"
)

func PlansHandler(w http.ResponseWriter, r *http.Request) {
	plans, err := PlanRepo.GetAllPlans()
	if err != nil {
		http.Error(w, "Failed to load investment plans", http.StatusInternalServerError)
		return
	}

	data := struct {
		Plans interface{}
	}{
		Plans: plans,
	}

	views.RenderTemplate(w, r, "plans.html", data)
}
