package services

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func ExtractTextFromImage(path string) (string, error) {
	tesseractPath := `C:\Program Files\Tesseract-OCR\tesseract.exe`
	cmd := exec.Command(tesseractPath, path, "stdout", "-l", "eng+mal")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("OCR failed: %v", err)
	}
	return string(out), nil
}

func DetectAmount(text string) float64 {
	// Clean text and look for "Total" or "Amount" lines
	re := regexp.MustCompile(`(?i)(?:total|sum|amount|net|payable|ആകെ|തുക)[:\s]*[^\d]*(\d+[\.,]\d{2}|\d+)`)
	matches := re.FindAllStringSubmatch(text, -1)

	if len(matches) > 0 {
		valStr := matches[len(matches)-1][1]
		valStr = strings.Replace(valStr, ",", ".", 1)
		amt, _ := strconv.ParseFloat(valStr, 64)
		return amt
	}

	// Fallback: Pick the largest decimal number on the bill
	reFallback := regexp.MustCompile(`\d+[\.,]\d{2}`)
	nums := reFallback.FindAllString(text, -1)
	var max float64
	for _, n := range nums {
		v, _ := strconv.ParseFloat(strings.Replace(n, ",", ".", 1), 64)
		if v > max { max = v }
	}
	return max
}