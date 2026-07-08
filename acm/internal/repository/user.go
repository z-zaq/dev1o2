package repository

import (
	"acm/internal/models"
	"database/sql"
	// "golang.org/x/tools/go/analysis/passes/nilfunc"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	is_admin BOOLEAN DEFAULT FALSE)`
	_, err := r.DB.Exec(query)
	return err
}
func (r *UserRepository) CreateUser(user models.User) error {
	query := `
	INSERT INTO users(name, email, password, is_admin)
	VALUES (?,?,?,?)`
	_, err := r.DB.Exec(
		query,
		user.Name,
		user.Email,
		user.Password,
		user.IsAdmin,
	)
	return err
}
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
	SELECT id, name, email, password, is_admin
	FROM users
	WHERE email = ?`

	user := &models.User{}

	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	rows, err := r.DB.Query(`
	SELECT id, name, email
	FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
		)
		if err != nil {
			return users, nil
		}
		users = append(users, user)
	}
	return users, nil
}
