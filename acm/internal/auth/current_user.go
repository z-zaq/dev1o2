package auth

import (
	"acm/internal/models"
	"acm/internal/repository"
	"net/http"
)

func GetCurrentUser(
	r *http.Request,
	userRepo *repository.UserRepository,
) (*models.User, error) {

	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	email, exists := Sessions[cookie.Value]
	if !exists {
		return nil, err
	}

	return userRepo.GetUserByEmail(email)
}
