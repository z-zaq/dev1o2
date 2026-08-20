package models

import "time"

type Investment struct {
	ID        int
	UserID    int
	PlanID    int
	Principal float64
	StartedAt time.Time
	EndsAt    time.Time
	Status    string // "active", "matured", or "withdrawn"
}
