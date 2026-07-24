package repository

import (
	"acm/internal/models"
	"database/sql"
	"strings"
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
	role TEXT NOT NULL DEFAULT 'user')`
	if _, err := r.DB.Exec(query); err != nil {
		return err
	}

	// Migration path for databases created before this change: add the
	// role column if it's not there yet, then backfill it from the old
	// is_admin flag so existing admin accounts don't lose their access.
	_, err := r.DB.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}
	// Best-effort backfill; ignore the error if is_admin never existed
	// (e.g. on a brand-new database that never had the old column).
	r.DB.Exec(`UPDATE users SET role = 'admin' WHERE is_admin = 1 AND role != 'admin'`)

	return nil
}

func (r *UserRepository) CreateUser(user models.User) error {
	query := `
	INSERT INTO users(name, email, password, role)
	VALUES (?,?,?,?)`
	_, err := r.DB.Exec(
		query,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
	)
	return err
}
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
	SELECT id, name, email, password, role
	FROM users
	WHERE email = ?`

	user := &models.User{}

	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	rows, err := r.DB.Query(`
	SELECT id, name, email, role
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
			&user.Role,
		)
		if err != nil {
			return users, nil
		}
		users = append(users, user)
	}
	return users, nil
}
func (r *UserRepository) UpdateUser(user models.User) error {
	query := `
	UPDATE users
	SET name = ?, email = ?
	WHERE id = ?
	`

	_, err := r.DB.Exec(
		query,
		user.Name,
		user.Email,
		user.ID,
	)

	return err
}
func (r *UserRepository) DeleteUser(userID int) error {
	query := `
	DELETE FROM users
	WHERE id = ?
	`

	_, err := r.DB.Exec(query, userID)
	return err
}
func (r *UserRepository) UpdatePassword(
	userID int,
	hashedPassword string,
) error {

	query := `
	UPDATE users
	SET password = ?
	WHERE id = ?
	`

	_, err := r.DB.Exec(
		query,
		hashedPassword,
		userID,
	)

	return err
}
