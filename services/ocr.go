package services

import (
	"os/exec"
	"regexp"
	"strconv"
)

func ExtractTextFromImage(path string) (string, error) {
	cmd := exec.Command("tesseract", path, "stdout")
	out, err := cmd.Output()
	return string(out), err
}

func DetectAmount(text string) float64 {
	// Look for patterns like "Total: 500" or "Amount 500.00"
	re := regexp.MustCompile(`(?i)(?:total|amount|net|sum|paid)[:\s]*[^\d]*(\d+(?:\.\d{2})?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	if len(matches) > 0 {
		// Take the last match (usually totals are at the bottom)
		lastMatch := matches[len(matches)-1][1]
		amt, _ := strconv.ParseFloat(lastMatch, 64)
		return amt
	}
	return 0
}