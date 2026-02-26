package services

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Convert WhatsApp voice (ogg) -> text using Whisper API
func SpeechToText(filePath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// add file
	part, _ := writer.CreateFormFile("file", "audio.ogg")
	io.Copy(part, file)

	// model
	writer.WriteField("model", "whisper-1")
	writer.Close()

	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/audio/transcriptions",
		body)

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	text, _ := result["text"].(string)
	return text, nil
}
