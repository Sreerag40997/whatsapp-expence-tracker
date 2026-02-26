package services

import (
	"context"
	"os"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func AppendRow(text string) error {
	ctx := context.Background()

	credFile := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	sheetID := os.Getenv("SHEET_ID")

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		return err
	}

	date := time.Now().Format("02-01-2006 15:04")
	values := [][]interface{}{
		{date, text},
	}

	_, err = srv.Spreadsheets.Values.Append(
		sheetID,
		"Sheet1!A:B",
		&sheets.ValueRange{Values: values},
	).ValueInputOption("RAW").Do()

	return err
}
