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
	// WhatsApp sends .ogg (opus). Whisper prefers .mp3 or .wav
	wavPath := audioPath + ".wav"
	
	// Ensure ffmpeg is installed and in your PATH
	cmd := exec.Command("ffmpeg", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", wavPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v", err)
	}
	defer os.Remove(wavPath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	file, err := os.Open(wavPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, _ := writer.CreateFormFile("file", filepath.Base(wavPath))
	io.Copy(part, file)
	
	writer.WriteField("model", "whisper-1")
	// Whisper auto-detects Malayalam, but you can hint it:
	writer.WriteField("language", "ml") 
	writer.Close()

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Text string `json:"text"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Text, nil
}