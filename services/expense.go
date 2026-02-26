package services

import (
	"bot/models"
	"sync"
	"time"
)

var (
	expenses []models.Expense
	mu       sync.Mutex
)

func AddExpense(amount float64, note string) {
	mu.Lock()
	defer mu.Unlock()

	expenses = append(expenses, models.Expense{
		Amount: amount,
		Note:   note,
		Date:   time.Now(),
	})

	// Also sync to Google Sheets
	AppendExpenseToSheet(note, amount)
}

func GetTotalExpense() float64 {
	mu.Lock()
	defer mu.Unlock()

	var total float64
	for _, e := range expenses {
		total += e.Amount
	}
	return total
}

func GetAllExpenses() []models.Expense {
	mu.Lock()
	defer mu.Unlock()
	return expenses
}