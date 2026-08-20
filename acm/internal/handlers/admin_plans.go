package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"acm/internal/models"
	"acm/internal/views"
)

func AdminCreatePlanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		views.RenderTemplate(w, r, "admin_plan.html", nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	assetClass := strings.TrimSpace(r.FormValue("asset_class"))
	rateStructure := strings.TrimSpace(r.FormValue("rate_structure"))

	duration, err := strconv.Atoi(r.FormValue("duration"))
	if err != nil || duration <= 0 {
		http.Error(w, "Duration must be a positive number", http.StatusBadRequest)
		return
	}

	minDeposit, err := strconv.ParseFloat(r.FormValue("min_deposit"), 64)
	if err != nil || minDeposit <= 0 {
		http.Error(w, "Minimum deposit must be greater than zero", http.StatusBadRequest)
		return
	}

	maxDeposit, err := strconv.ParseFloat(r.FormValue("max_deposit"), 64)
	if err != nil || maxDeposit <= 0 {
		http.Error(w, "Maximum deposit must be greater than zero", http.StatusBadRequest)
		return
	}

	if name == "" || assetClass == "" || rateStructure == "" {
		http.Error(w, "All plan fields are required", http.StatusBadRequest)
		return
	}

	if maxDeposit < minDeposit {
		http.Error(w, "Maximum deposit cannot be less than minimum deposit", http.StatusBadRequest)
		return
	}

	plan := models.Plan{
		Name:          name,
		AssetClass:    assetClass,
		Duration:      duration,
		RateStructure: rateStructure,
		MinDeposit:    minDeposit,
		MaxDeposit:    maxDeposit,
	}

	if err := PlanRepo.CreatePlan(plan); err != nil {
		http.Error(w, "Failed to create investment plan", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/plans", http.StatusSeeOther)
}
