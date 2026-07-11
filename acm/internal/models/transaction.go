package models

import (
	"time"
)

type Transaction struct {
	ID        int
	UserID    int
	Type      string
	Amount    float64
	CreatedAt time.Time
}
