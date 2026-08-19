package middleware

import (
	"acm/internal/auth"
	"acm/internal/models"
	"acm/internal/repository"
	"context"
	"net/http"
)

type ctxKey string

const userCtxKey ctxKey = "currentUser"

// RequireAuth resolves the session cookie into a *models.User and attaches
// it to the request context before calling next. Redirects to /login if
// there's no valid (or non-expired) session.
func RequireAuth(userRepo *repository.UserRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		email, exists := auth.GetSessionEmail(cookie.Value)
		if !exists {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		user, err := userRepo.GetUserByEmail(email)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, user)
		next(w, r.WithContext(ctx))
	}
}

// RequireRole wraps RequireAuth and additionally checks the resolved
// user's role, returning 403 if it doesn't match.
func RequireRole(userRepo *repository.UserRepository, role string, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(userRepo, func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r)
		if user == nil || user.Role != role {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}

// UserFromContext retrieves the user attached by RequireAuth/RequireRole.
// Returns nil if called on a request that wasn't wrapped by either.
func UserFromContext(r *http.Request) *models.User {
	user, _ := r.Context().Value(userCtxKey).(*models.User)
	return user
}
