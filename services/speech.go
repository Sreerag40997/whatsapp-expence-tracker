package services

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func SpeechToText(audioPath string) (string, error) {
	wavPath := audioPath + ".wav"
	// Convert OGG to WAV
	exec.Command("ffmpeg", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", wavPath).Run()
	defer os.Remove(wavPath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, _ := os.Open(wavPath)
	part, _ := writer.CreateFormFile("file", filepath.Base(wavPath))
	io.Copy(part, file)
	file.Close()

	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "ml") // Hinting Malayalam
	writer.Close()

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, _ := http.DefaultClient.Do(req)
	var res struct {
		Text string `json:"text"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Text, nil
}
