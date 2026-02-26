package handlers

import (
	"bot/services"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func VerifyWebhook(c *gin.Context) {
	verifyToken := os.Getenv("VERIFY_TOKEN")
	if c.Query("hub.mode") == "subscribe" && c.Query("hub.verify_token") == verifyToken {
		c.String(200, c.Query("hub.challenge"))
		return
	}
	c.Status(403)
}

func ReceiveMessage(c *gin.Context) {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.Status(200)
		return
	}

	// Parse WhatsApp JSON structure
	entry, ok := body["entry"].([]interface{})
	if !ok || len(entry) == 0 {
		c.Status(200)
		return
	}
	changes := entry[0].(map[string]interface{})["changes"].([]interface{})
	value := changes[0].(map[string]interface{})["value"].(map[string]interface{})
	
	messages, exists := value["messages"].([]interface{})
	if !exists || len(messages) == 0 {
		c.Status(200)
		return
	}

	msg := messages[0].(map[string]interface{})
	from := msg["from"].(string)
	msgType := msg["type"].(string)

	switch msgType {
	case "text":
		text := msg["text"].(map[string]interface{})["body"].(string)
		handleText(from, text)
	case "image":
		image := msg["image"].(map[string]interface{})
		path, _ := services.DownloadWhatsAppMedia(image["id"].(string))
		text, _ := services.ExtractTextFromImage(path)
		amt := services.DetectAmount(text)
		if amt > 0 {
			services.AddExpense(amt, "Bill OCR")
			sendMessage(from, fmt.Sprintf("✅ Added from Image: ₹%.2f", amt))
		} else {
			sendMessage(from, "❌ Could not detect amount on bill.")
		}
	case "audio":
		audio := msg["audio"].(map[string]interface{})
		path, _ := services.DownloadWhatsAppMedia(audio["id"].(string))
		text, _ := services.SpeechToText(path)
		note, amt, ok := services.ParseExpense(text)
		if ok {
			services.AddExpense(amt, note)
			sendMessage(from, fmt.Sprintf("🎤 Voice Added: %s - ₹%.2f", note, amt))
		}
	}

	c.Status(200)
}

func handleText(from, text string) {
	cleanText := strings.ToLower(strings.TrimSpace(text))

	if cleanText == "hi" || cleanText == "hello" {
		sendMessage(from, "👋 Welcome! Send 'Lunch 200', a bill photo, or '/expenses'.")
		return
	}

	if cleanText == "/expenses" {
		total := services.GetTotalExpense()
		sendMessage(from, fmt.Sprintf("💰 Total: ₹%.2f", total))
		return
	}

	if cleanText == "/statement" {
		file := services.GenerateMonthlyPDF()
		sendDocument(from, file)
		return
	}

	note, amt, ok := services.ParseExpense(text)
	if ok {
		services.AddExpense(amt, note)
		sendMessage(from, fmt.Sprintf("✅ Added: %s - ₹%.2f", note, amt))
	} else {
		sendMessage(from, "❌ Try: 'Food 500'")
	}
}

func sendMessage(to, text string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]string{"body": text},
	}
	sendRequest(url, payload)
}

func sendDocument(to, fileName string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	fileURL := os.Getenv("BASE_URL") + "/public/" + fileName
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "document",
		"document": map[string]string{
			"link":     fileURL,
			"filename": "Statement.txt",
		},
	}
	sendRequest(url, payload)
}

func sendRequest(url string, payload interface{}) {
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(data)))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}