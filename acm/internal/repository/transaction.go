package repository

import (
	"database/sql"
)

type TransactionRepository struct {
	DB *sql.DB
}

func (r *TransactionRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		amount REAL NOT NULL
	);`
	_, err := r.DB.Exec(query)
	return err
}
