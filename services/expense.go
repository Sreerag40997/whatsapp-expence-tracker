package services

import (
	"bot/models"
	"fmt"
	"sync"
	"time"
)

var (
	expenses []models.Expense
	limit    float64 = 0
	mu       sync.Mutex
)

func AddExpense(amount float64, note string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	expenses = append(expenses, models.Expense{
		Amount: amount, Note: note, Date: time.Now(),
	})

	// Log to Google Sheets (Background)
	go AppendExpenseToSheet(note, amount)

	total := 0.0
	for _, e := range expenses {
		total += e.Amount
	}

	if limit > 0 && total > limit {
		return fmt.Sprintf("⚠️ *BUDGET WARNING*\nYou have crossed your limit of ₹%.2f! Total: ₹%.2f", limit, total), true
	}
	return "", false
}

func GetMonthlySummary(month int, year int) string {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	if month == 0 {
		month = int(now.Month())
	}
	if year == 0 {
		year = now.Year()
	}

	var total float64
	itemsText := ""
	for _, e := range expenses {
		if int(e.Date.Month()) == month && e.Date.Year() == year {
			itemsText += fmt.Sprintf("📍 *%-12s* : ₹%.2f\n", e.Note, e.Amount)
			total += e.Amount
		}
	}

	if itemsText == "" {
		return "ℹ️ *No expenses recorded.*"
	}

	bill := fmt.Sprintf("🧾 *STATEMENT (%02d/%d)*\n━━━━━━━━━━━━━━━━━━━━\n", month, year)
	bill += itemsText + "━━━━━━━━━━━━━━━━━━━━\n"
	bill += fmt.Sprintf("💰 *TOTAL: ₹%.2f*", total)
	return bill
}

func SetLimit(amt float64) { mu.Lock(); limit = amt; mu.Unlock() }
func ResetExpenses()       { mu.Lock(); expenses = []models.Expense{}; limit = 0; mu.Unlock() }
