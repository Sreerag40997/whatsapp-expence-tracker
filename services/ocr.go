package services

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func ExtractTextFromImage(path string) (string, error) {
	cmd := exec.Command("tesseract", path, "stdout")
	out, err := cmd.Output()
	return string(out), err
}

func DetectAmount(text string) float64 {
	// Cleans commas and looks for digits after "Total/Amount/Sum"
	re := regexp.MustCompile(`(?i)(?:total|sum|amt|amount|net|paid)[:\s]*[^\d]*(\d+[\.,]\d{2}|\d+)`)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) > 0 {
		valStr := matches[len(matches)-1][1]
		valStr = strings.Replace(valStr, ",", ".", 1)
		amt, _ := strconv.ParseFloat(valStr, 64)
		return amt
	}
	return 0
}
