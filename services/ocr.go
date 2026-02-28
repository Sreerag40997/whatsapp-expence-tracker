package services

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func ExtractTextFromImage(path string) (string, error) {
	// Point to the exact location of Tesseract on your Windows machine
	tesseractPath := `C:\Program Files\Tesseract-OCR\tesseract.exe`
	
	// Use "eng" and "mal" (if you installed Malayalam data) for better results
	cmd := exec.Command(tesseractPath, path, "stdout", "-l", "eng+mal")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func DetectAmount(text string) float64 {
	// Improved Regex: Looks for Total/Amount followed by numbers
	// Handles common variations and Malayalam symbols if any
	re := regexp.MustCompile(`(?i)(?:total|amount|sum|net|payable|grand|balance|ആകെ|തുക)[:\s]*[^\d]*(\d+[\.,]\d{2}|\d+)`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	if len(matches) > 0 {
		// Take the last match (usually the final total at bottom of bill)
		valStr := matches[len(matches)-1][1]
		valStr = strings.Replace(valStr, ",", ".", 1)
		amt, _ := strconv.ParseFloat(valStr, 64)
		return amt
	}
	
	// Fallback: If no "Total" keyword found, look for any large number with decimals
	reFallback := regexp.MustCompile(`(\d+[\.,]\d{2})`)
	fallbacks := reFallback.FindAllString(text, -1)
	if len(fallbacks) > 0 {
		valStr := strings.Replace(fallbacks[len(fallbacks)-1], ",", ".", 1)
		amt, _ := strconv.ParseFloat(valStr, 64)
		return amt
	}

	return 0
}