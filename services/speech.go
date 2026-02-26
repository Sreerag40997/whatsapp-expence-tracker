package services

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func SpeechToText(audioPath string) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: audioPath,
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
