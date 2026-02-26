package services

import (
	"strconv"
	"strings"
)

func ParseExpenseText(text string) (bool, string, float64) {
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		return false, "", 0
	}

	title := parts[0]
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return false, "", 0
	}

	return true, title, amount
}
