package services

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func ParseExpense(text string) (string, float64, bool) {
	// 1. Clean the string
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "രൂപ", "")
	text = strings.ReplaceAll(text, "rupees", "")
	text = strings.ReplaceAll(text, "rs", "")
	text = strings.ReplaceAll(text, "₹", "")

	// 2. Extract the Number (Amount)
	reNum := regexp.MustCompile(`(\d+(\.\d+)?)`)
	matchNum := reNum.FindString(text) // Fixed the undefined compiler error

	if matchNum == "" {
		return "", 0, false
	}

	amt, _ := strconv.ParseFloat(matchNum, 64)

	// 3. Extract the Item (Note)
	// Remove the number and common connector words
	notePart := reNum.ReplaceAllString(text, "")
	notePart = strings.ReplaceAll(notePart, "for", "")
	notePart = strings.ReplaceAll(notePart, "spent", "")

	note := strings.TrimSpace(notePart)
	if note == "" {
		note = "Miscellaneous Expense"
	}

	// Capitalize first letter for a premium look
	r := []rune(note)
	r[0] = unicode.ToUpper(r[0])

	return string(r), amt, true
}
