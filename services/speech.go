package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func SpeechToText(audioPath string) (string, error) {
	// convert to wav using ffmpeg
	wavPath := "tmp/audio.wav"
	cmd := exec.Command("C:\\ffmpeg\\bin\\ffmpeg.exe", "-y", "-i", audioPath, wavPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// OpenAI Whisper API
	apiKey := os.Getenv("OPENAI_API_KEY")

	file, _ := os.Open(wavPath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := io.MultiWriter(body)
	io.Copy(writer, file)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "audio/wav")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result["text"].(string), nil
}
