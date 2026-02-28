package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func SpeechToText(audioPath string) (string, error) {
	wavPath := audioPath + ".wav"

	// 1. Convert OGG to WAV using FFmpeg
	// -y overwrites, -ar 16000 is optimal for Whisper
	cmd := exec.Command("ffmpeg", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", wavPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("FFMPEG_ERROR")
	}
	defer os.Remove(wavPath)

	// 2. Prepare Multi-part request for OpenAI
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := os.Open(wavPath)
	if err != nil {
		return "", err
	}
	part, _ := writer.CreateFormFile("file", filepath.Base(wavPath))
	io.Copy(part, file)
	file.Close()

	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "ml") // Malayalam support
	writer.Close()

	// 3. Send to OpenAI
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 4. Handle Quota/Billing Errors (Error 429)
	if resp.StatusCode == 429 {
		return "", fmt.Errorf("OPENAI_QUOTA_EXCEEDED")
	}

	var result struct {
		Text string `json:"text"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Text == "" {
		return "", fmt.Errorf("EMPTY_TRANSCRIPTION")
	}

	return result.Text, nil
}
