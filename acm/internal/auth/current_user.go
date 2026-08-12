package auth

import (
	"acm/internal/models"
	"acm/internal/repository"
	"errors"
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
		return nil, errors.New("session not found or expired")
	}

	return userRepo.GetUserByEmail(email)
}
