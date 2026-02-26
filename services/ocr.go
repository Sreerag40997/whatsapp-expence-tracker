package services

import "os/exec"

func ExtractTextFromImage(path string) (string, error) {
	cmd := exec.Command("tesseract", path, "stdout")
	out, err := cmd.Output()
	return string(out), err
}

// func SpeechToText(audioPath string) (string, error) {
// 	// Mock: In production, use OpenAI Whisper API here
// 	return "Coffee 150", nil
// }