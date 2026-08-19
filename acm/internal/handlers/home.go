package handlers

import (
	"acm/internal/views"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTemplate(w, r, "home.html", nil)
}
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTemplate(w, r, "about.html", nil)
}
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTemplate(w, r, "contact.html", nil)
}
