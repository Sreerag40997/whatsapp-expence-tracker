package services

import (
	"fmt"
	"os"
	"time"
)

func GenerateMonthlyPDF() string {
	fileName := fmt.Sprintf("statement_%d_%d.txt", time.Now().Month(), time.Now().Year())

	content := fmt.Sprintf("Total Expense: ₹%.2f\n", GetTotalExpense())

	os.WriteFile(fileName, []byte(content), 0644)

	return fileName
}
