package handlers

import (
	"acm/internal/views"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTemplate(w, "dashboard.html")
}
