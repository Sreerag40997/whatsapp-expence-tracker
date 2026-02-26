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
	
	// Step 1: Get Media URL
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", mediaID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get media metadata")
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	// Step 2: Download the actual file
	req, _ = http.NewRequest("GET", result.URL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download file")
	}
	defer resp.Body.Close()

	os.MkdirAll("tmp", 0755)
	filePath := filepath.Join("tmp", mediaID+".jpg")
	file, _ := os.Create(filePath)
	defer file.Close()

	io.Copy(file, resp.Body)
	return filePath, nil
}