package services

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseExpense -> "Lunch 200" , "food 300"
func ParseExpense(text string) (string, float64, bool) {
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	// regex to find amount
	re := regexp.MustCompile(`(\d+(\.\d+)?)`)
	match := re.FindString(text)

	if match == "" {
		return "", 0, false
	}

	amt, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return "", 0, false
	}

	// remove amount from string to get note
	note := strings.Replace(text, match, "", 1)
	note = strings.TrimSpace(note)

	if note == "" {
		note = "Expense"
	}

	return note, amt, true
}

func DetectAmount(text string) float64 {
	text = strings.ToLower(text)

	re := regexp.MustCompile(`(\d+(\.\d{1,2})?)`)
	matches := re.FindAllString(text, -1)

	if len(matches) == 0 {
		return 0
	}

	last := matches[len(matches)-1]
	amt, err := strconv.ParseFloat(last, 64)
	if err != nil {
		return 0
	}

	return amt
}
