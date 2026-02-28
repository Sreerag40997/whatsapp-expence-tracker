package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadWhatsAppMedia(mediaID string) (string, error) {
	token := os.Getenv("ACCESS_TOKEN")

	// STEP 1: Get the Media Metadata (The actual URL)
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", mediaID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.URL == "" {
		return "", fmt.Errorf("failed to get media URL from Meta")
	}

	// STEP 2: Download the actual File Bytes
	reqDown, _ := http.NewRequest("GET", result.URL, nil)
	reqDown.Header.Set("Authorization", "Bearer "+token)
	// IMPORTANT: Meta sometimes blocks requests without a User-Agent
	reqDown.Header.Set("User-Agent", "Mozilla/5.0")

	respDown, err := http.DefaultClient.Do(reqDown)
	if err != nil {
		return "", err
	}
	defer respDown.Body.Close()

	// Ensure tmp directory exists
	os.MkdirAll("tmp", 0755)

	// Create file path (Adding .jpg extension helps Windows/Tesseract)
	localPath := filepath.Join("tmp", mediaID+".jpg")
	file, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy bytes to file
	size, err := io.Copy(file, respDown.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("📂 Media Downloaded: %s (Size: %d bytes)\n", localPath, size)
	return localPath, nil
}
