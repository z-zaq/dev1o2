package repository

import (
	"database/sql"
	"acm/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL);`
	_, err := r.DB.Exec(query)
	return err
}
func (r *UserRepository) CreateUser(user models.User) error{
	query := `
	INSERT INTO users(name, email, password)
	VALUES (?,?,?)`
	_, err := r.DB.Exec(
		query,
		user.Name,
		user.Email,
		user.Password,
	)
	return err
}
