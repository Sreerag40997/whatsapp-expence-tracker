package services

// import (
// 	"fmt"
// 	"time"

// 	"github.com/jung-kurt/gofpdf"
// )

// func GenerateMonthlyPDF() string {
// 	fileName := fmt.Sprintf("statement_%d.pdf", time.Now().Unix())
// 	filePath := "public/" + fileName

// 	pdf := gofpdf.New("P", "mm", "A4", "")
// 	pdf.AddPage()
// 	pdf.SetFont("Arial", "B", 16)
// 	pdf.Cell(40, 10, "Monthly Expense Statement")

// 	expenses := GetAllExpenses()
// 	pdf.Ln(10)

// 	for _, e := range expenses {
// 		line := fmt.Sprintf("%s - ₹%.2f", e.Note, e.Amount)
// 		pdf.Cell(40, 10, line)
// 		pdf.Ln(8)
// 	}

// 	pdf.OutputFileAndClose(filePath)
// 	return fileName
// }
