package repository

import (
	"acm/internal/models"
	"database/sql"
	"strings"
)

type PlanRepository struct {
	DB *sql.DB
}

func (r *PlanRepository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS plans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		asset_class TEXT NOT NULL,
		duration INTEGER NOT NULL,
		rate_type TEXT NOT NULL DEFAULT 'fixed_compounding',
		rate_value REAL NOT NULL DEFAULT 0,
		min_deposit REAL NOT NULL,
		max_deposit REAL NOT NULL
	)`

	if _, err := r.DB.Exec(query); err != nil {
		return err
	}

	if _, err := r.DB.Exec(`ALTER TABLE plans ADD COLUMN rate_type TEXT NOT NULL DEFAULT 'fixed_compounding'`); err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}
	if _, err := r.DB.Exec(`ALTER TABLE plans ADD COLUMN rate_value REAL NOT NULL DEFAULT 0`); err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}
	r.DB.Exec(`UPDATE plans SET rate_value = CAST(rate_structure AS REAL) WHERE rate_value = 0`)

	return nil
}

func (r *PlanRepository) CreatePlan(plan models.Plan) error {
	query := `
	INSERT INTO plans (
		name,
		asset_class,
		duration,
		rate_type,
		rate_value,
		min_deposit,
		max_deposit
	)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.DB.Exec(
		query,
		plan.Name,
		plan.AssetClass,
		plan.Duration,
		plan.RateType,
		plan.RateValue,
		plan.MinDeposit,
		plan.MaxDeposit,
	)

	return err
}

func (r *PlanRepository) GetPlanByID(planID int) (*models.Plan, error) {
	query := `
	SELECT
		id,
		name,
		asset_class,
		duration,
		rate_type,
		rate_value,
		min_deposit,
		max_deposit
	FROM plans
	WHERE id = ?`

	plan := &models.Plan{}

	err := r.DB.QueryRow(query, planID).Scan(
		&plan.ID,
		&plan.Name,
		&plan.AssetClass,
		&plan.Duration,
		&plan.RateType,
		&plan.RateValue,
		&plan.MinDeposit,
		&plan.MaxDeposit,
	)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (r *PlanRepository) GetAllPlans() ([]models.Plan, error) {
	query := `
	SELECT
		id,
		name,
		asset_class,
		duration,
		rate_type,
		rate_value,
		min_deposit,
		max_deposit
	FROM plans
	ORDER BY id`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []models.Plan

	for rows.Next() {
		var plan models.Plan

		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.AssetClass,
			&plan.Duration,
			&plan.RateType,
			&plan.RateValue,
			&plan.MinDeposit,
			&plan.MaxDeposit,
		)
		if err != nil {
			return nil, err
		}

		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}
