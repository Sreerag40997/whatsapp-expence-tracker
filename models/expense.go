package models

import "time"

type Expense struct {
	ID     string    `json:"id"`
	Amount float64   `json:"amount"`
	Note   string    `json:"note"`
	Date   time.Time `json:"date"`
	Month  int       `json:"month"`
	Year   int       `json:"year"`
}
