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
		Amount: amount,
		Note:   note,
		Date:   time.Now(),
	})

	total := 0.0
	for _, e := range expenses {
		total += e.Amount
	}

	if limit > 0 && total > limit {
		return fmt.Sprintf("⚠️ *LIMIT EXCEEDED!*\nLimit: ₹%.2f\nTotal: ₹%.2f", limit, total), true
	}
	return "", false
}

func GetMonthlySummary(month int, year int) string {
	mu.Lock()
	defer mu.Unlock()

	// Default to current month/year if not provided
	now := time.Now()
	if month == 0 { month = int(now.Month()) }
	if year == 0 { year = now.Year() }

	var total float64
	itemsText := ""
	
	for _, e := range expenses {
		if int(e.Date.Month()) == month && e.Date.Year() == year {
			itemsText += fmt.Sprintf("• %s: ₹%.2f\n", e.Note, e.Amount)
			total += e.Amount
		}
	}

	if itemsText == "" {
		return fmt.Sprintf("📝 No expenses recorded for %02d/%d", month, year)
	}

	bill := fmt.Sprintf("🧾 *EXPENSE BILL (%02d/%d)*\n", month, year)
	bill += "--------------------------\n"
	bill += itemsText
	bill += "--------------------------\n"
	bill += fmt.Sprintf("💰 *TOTAL: ₹%.2f*\n", total)
	
	if limit > 0 {
		bill += fmt.Sprintf("📊 *LIMIT: ₹%.2f*\n", limit)
		if total > limit {
			bill += fmt.Sprintf("⚠️ *OVER BY: ₹%.2f*", total-limit)
		} else {
			bill += fmt.Sprintf("✅ *REMAINING: ₹%.2f*", limit-total)
		}
	}
	return bill
}

func GetTotalExpense() float64 {
	mu.Lock()
	defer mu.Unlock()
	var total float64
	for _, e := range expenses { total += e.Amount }
	return total
}

func SetLimit(amt float64) { limit = amt }
func ResetExpenses() { mu.Lock(); expenses = []models.Expense{}; mu.Unlock() }