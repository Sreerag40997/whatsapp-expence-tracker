package services

import (
	"os/exec"
)

func ExtractText(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
