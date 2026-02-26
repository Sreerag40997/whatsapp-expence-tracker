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

// ➕ Add Expense
func AddExpense(amount float64, note string) {
	mu.Lock()
	defer mu.Unlock()

	expenses = append(expenses, models.Expense{
		Amount: amount,
		Note:   note,
		Date:   time.Now(),
	})
}

// 💰 Get Total Expense
func GetTotalExpense() float64 {
	mu.Lock()
	defer mu.Unlock()

	var total float64
	for _, e := range expenses {
		total += e.Amount
	}
	return total
}

// 📄 Get All Expenses (Safe Copy)
func GetAllExpenses() []models.Expense {
	mu.Lock()
	defer mu.Unlock()

	// important: return copy to avoid race condition
	copySlice := make([]models.Expense, len(expenses))
	copy(copySlice, expenses)
	return copySlice
}

// 🧹 Reset Monthly (for month end cron later)
func ResetExpenses() {
	mu.Lock()
	defer mu.Unlock()
	expenses = []models.Expense{}
}
