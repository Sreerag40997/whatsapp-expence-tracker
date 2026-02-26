package services

import (
	"regexp"
	"strconv"
	"strings"
)

func ParseExpense(text string) (string, float64, bool) {
	text = strings.ToLower(strings.TrimSpace(text))

	re := regexp.MustCompile(`([a-zA-Z ]+)\s*(\d+)`)
	match := re.FindStringSubmatch(text)

	if len(match) < 3 {
		return "", 0, false
	}

	title := strings.TrimSpace(match[1])
	amt, _ := strconv.ParseFloat(match[2], 64)

	return title, amt, true
}
