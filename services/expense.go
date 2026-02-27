package services

import (
	"bot/models"
	"fmt"
	"sync"
	"time"
)

var (
	expenses []models.Expense
	limit    float64 = 0 // 0 means no limit set
	mu       sync.Mutex
)

func AddExpense(amount float64, note string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	expenses = append(expenses, models.Expense{
		Amount: amount,
		Note:   note,
		Date:   time.Now(),
	})

	total := 0.0
	for _, e := range expenses {
		total += e.Amount
	}

	// Check Limit
	if limit > 0 && total > limit {
		warning := fmt.Sprintf("⚠️ WARNING: You have exceeded your limit of ₹%.2f! (Current Total: ₹%.2f)", limit, total)
		return warning, true
	}
	return "", false
}

func SetLimit(amt float64) {
	mu.Lock()
	defer mu.Unlock()
	limit = amt
}

func ResetExpenses() {
	mu.Lock()
	defer mu.Unlock()
	expenses = []models.Expense{}
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