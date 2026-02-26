package services

import (
	"fmt"
	"os"
	"time"
)

func GenerateMonthlyPDF() string {
	fileName := fmt.Sprintf("statement_%d.txt", time.Now().Unix()) 
	content := "--- Monthly Expense Statement ---\n"
	for _, e := range GetAllExpenses() {
		content += fmt.Sprintf("%s: %s - ₹%.2f\n", e.Date.Format("2006-01-02"), e.Note, e.Amount)
	}

	os.MkdirAll("public", 0755)
	os.WriteFile("public/"+fileName, []byte(content), 0644)
	return fileName
}