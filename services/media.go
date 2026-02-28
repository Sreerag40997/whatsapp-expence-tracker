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
	
	// STEP 1: Get the Download URL from Meta
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
		return "", fmt.Errorf("failed to get media URL")
	}

	// STEP 2: Download the actual file bytes
	reqDown, _ := http.NewRequest("GET", result.URL, nil)
	reqDown.Header.Set("Authorization", "Bearer "+token)
	
	respDown, err := http.DefaultClient.Do(reqDown)
	if err != nil {
		return "", err
	}
	defer respDown.Body.Close()

	os.MkdirAll("tmp", 0755)
	filePath := filepath.Join("tmp", mediaID) 
	
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, respDown.Body)
	return filePath, err
}