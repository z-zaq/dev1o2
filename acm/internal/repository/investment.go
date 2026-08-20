package repository

import (
	"acm/internal/models"
	"database/sql"
	"time"
)

type InvestmentRepository struct {
	DB *sql.DB
}

func (r *InvestmentRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS investments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		plan_id INTEGER NOT NULL,
		principal REAL NOT NULL,
		started_at DATETIME NOT NULL,
		ends_at DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'active'
	)`
	_, err := r.DB.Exec(query)
	return err
}

func (r *InvestmentRepository) CreateInvestment(
	investment models.Investment,
) error {
	query := `
	INSERT INTO investments (
		user_id,
		plan_id,
		principal,
		started_at,
		ends_at,
		status
	)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.DB.Exec(
		query,
		investment.UserID,
		investment.PlanID,
		investment.Principal,
		investment.StartedAt,
		investment.EndsAt,
		investment.Status,
	)

	return err
}

func (r *InvestmentRepository) GetInvestmentByID(
	investmentID int,
) (*models.Investment, error) {
	query := `
	SELECT id, user_id, plan_id, principal, started_at, ends_at, status
	FROM investments
	WHERE id = ?`

	investment := &models.Investment{}

	err := r.DB.QueryRow(query, investmentID).Scan(
		&investment.ID,
		&investment.UserID,
		&investment.PlanID,
		&investment.Principal,
		&investment.StartedAt,
		&investment.EndsAt,
		&investment.Status,
	)

	if err != nil {
		return nil, err
	}

	return investment, nil
}

func (r *InvestmentRepository) GetInvestmentsByUserID(
	userID int,
) ([]models.Investment, error) {
	query := `
	SELECT id, user_id, plan_id, principal, started_at, ends_at, status
	FROM investments
	WHERE user_id = ?
	ORDER BY started_at DESC`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investments []models.Investment

	for rows.Next() {
		var investment models.Investment

		err := rows.Scan(
			&investment.ID,
			&investment.UserID,
			&investment.PlanID,
			&investment.Principal,
			&investment.StartedAt,
			&investment.EndsAt,
			&investment.Status,
		)
		if err != nil {
			return nil, err
		}

		investments = append(investments, investment)
	}

	return investments, nil
}

func (r *InvestmentRepository) GetAllInvestments() ([]models.Investment, error) {
	query := `
	SELECT id, user_id, plan_id, principal, started_at, ends_at, status
	FROM investments
	ORDER BY started_at DESC`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investments []models.Investment

	for rows.Next() {
		var investment models.Investment

		err := rows.Scan(
			&investment.ID,
			&investment.UserID,
			&investment.PlanID,
			&investment.Principal,
			&investment.StartedAt,
			&investment.EndsAt,
			&investment.Status,
		)
		if err != nil {
			return nil, err
		}

		investments = append(investments, investment)
	}

	return investments, nil
}

func (r *InvestmentRepository) UpdateStatus(
	investmentID int,
	status string,
) error {
	_, err := r.DB.Exec(
		`UPDATE investments SET status = ? WHERE id = ?`,
		status,
		investmentID,
	)

	return err
}

func (r *InvestmentRepository) CreateInvestmentWithFunding(
	investment models.Investment,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO investments (
			user_id,
			plan_id,
			principal,
			started_at,
			ends_at,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?)`,
		investment.UserID,
		investment.PlanID,
		investment.Principal,
		investment.StartedAt,
		investment.EndsAt,
		investment.Status,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (
			user_id,
			type,
			amount,
			description,
			created_at
		)
		VALUES (?, ?, ?, ?, ?)`,
		investment.UserID,
		"investment",
		investment.Principal,
		"Investment funding",
		investment.StartedAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *InvestmentRepository) MatureInvestment(
	investmentID int,
	userID int,
	principal float64,
	returnAmount float64,
) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec(`
		UPDATE investments
		SET status = 'matured'
		WHERE id = ?
		  AND user_id = ?
		  AND status = 'active'
	`, investmentID, userID)

	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return sql.ErrNoRows
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (
			user_id,
			type,
			amount,
			description,
			created_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		userID,
		"investment_return",
		principal+returnAmount,
		"Investment maturity payout",
		time.Now(),
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
