package models

import "time"

type Expense struct {
	Amount float64   `json:"amount"`
	Note   string    `json:"note"`
	Date   time.Time `json:"date"`
}