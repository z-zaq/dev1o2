package repository

import (
	"acm/internal/models"
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
	)`
	// query := `
	// CREATE TABLE IF NOT EXISTS transactions (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	name TEXT NOT NULL,
	// 	email TEXT UNIQUE NOT NULL,
	// 	password TEXT NOT NULL,
	// 	is_admin BOOLEAN DEFAULT FALSE
	// )`
	_, err := r.DB.Exec(query)
	return err
}
func (r *TransactionRepository) CreateTransaction(
	transaction models.Transaction,
) error {
	query := `
	INSERT INTO transactions(user_id, type, amount)
	VALUES (?, ?, ?)`
	_, err := r.DB.Exec(
		query,
		transaction.UserID,
		transaction.Type,
		transaction.Amount,
	)
	return err
}
func (r *TransactionRepository) GetTransactionsByUserID(
	userID int,
) ([]models.Transaction, error) {
	query := `
	SELECT id, user_id, type, amount
	FROM transactions
	WHERE user_id = ?`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction

	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Type,
			&transaction.Amount,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
func (r *TransactionRepository) GetBalanceByUserID(userID int) (float64, error) {
	query := `
	SELECT COALESCE(SUM(CASE 
	WHEN type = 'deposit' THEN amount
	WHEN type = 'withdraw' THEN -amount
	ELSE 0
	END), 0)
	FROM transactions
	WHERE user_id = ?
	`

	var balance float64

	err := r.DB.QueryRow(query, userID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
func (r *TransactionRepository) GetAllTransactions() ([]models.Transaction, error) {
	rows, err := r.DB.Query(`
	SELECT id, user_id, type, amount
	FROM transactions
	ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction

	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Type,
			&t.Amount,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
