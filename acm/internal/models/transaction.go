package models

import (
	"time"
)

type Transaction struct {
	ID          int
	UserID      int
	Type        string
	Amount      float64
	Description string
	CreatedAt   time.Time
}
