package services

import (
	"fmt"
	"os"
	"time"
)

func GenerateMonthlyPDF() string {
	fileName := fmt.Sprintf("statement_%d_%d.pdf", time.Now().Month(), time.Now().Year())

	content := "Monthly Expense Statement\n\n"
	for _, e := range GetAllExpenses() {
		content += fmt.Sprintf("%s - ₹%.2f - %s\n",
			e.Date.Format("02 Jan 2006"),
			e.Amount,
			e.Note)
	}

	os.MkdirAll("public", os.ModePerm)
	os.WriteFile("public/"+fileName, []byte(content), 0644)

	return fileName
}
