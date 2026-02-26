package services

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseExpense parses strings like "Lunch 200" or "200 food"
func ParseExpense(text string) (string, float64, bool) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return "", 0, false
	}

	// Check if last word is amount
	if amt, err := strconv.ParseFloat(parts[len(parts)-1], 64); err == nil {
		note := strings.Join(parts[:len(parts)-1], " ")
		return note, amt, true
	}

	// Check if first word is amount
	if amt, err := strconv.ParseFloat(parts[0], 64); err == nil {
		note := strings.Join(parts[1:], " ")
		return note, amt, true
	}

	return "", 0, false
}

// DetectAmount specifically for OCR text
func DetectAmount(text string) float64 {
	re := regexp.MustCompile(`\d+(\.\d+)?`)
	matches := re.FindAllString(text, -1)
	
	if len(matches) == 0 {
		return 0
	}

	var max float64
	for _, m := range matches {
		val, _ := strconv.ParseFloat(m, 64)
		if val > max {
			max = val
		}
	}
	return max
}