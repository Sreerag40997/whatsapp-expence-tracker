package services

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func AppendExpenseToSheet(note string, amount float64) {
	sheetID := os.Getenv("GOOGLE_SHEET_ID")
	ctx := context.Background()

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Println("Sheet Error:", err)
		return
	}

	values := [][]interface{}{{time.Now().Format("2006-01-02 15:04"), note, amount}}
	rb := &sheets.ValueRange{Values: values}

	_, err = srv.Spreadsheets.Values.Append(sheetID, "Sheet1!A:C", rb).
		ValueInputOption("RAW").Do()
	if err != nil {
		log.Println("Append Error:", err)
	}
}