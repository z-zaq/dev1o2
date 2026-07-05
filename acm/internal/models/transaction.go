package models

type Transaction struct {
	ID     int
	UserID int
	Type   string
	Amount float64
}
