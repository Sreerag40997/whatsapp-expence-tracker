package services

import (
	"io"
	"net/http"
	"os"
)

func DownloadWhatsAppMedia(mediaID string) string {
	url := "https://graph.facebook.com/v18.0/" + mediaID
	token := os.Getenv("ACCESS_TOKEN")

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	path := "tmp/media.jpg"
	file, _ := os.Create(path)
	defer file.Close()

	io.Copy(file, resp.Body)
	return path
}
