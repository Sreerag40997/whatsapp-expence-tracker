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
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == "mytoken" {
		c.String(200, challenge)
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

	entryArr, ok := body["entry"].([]interface{})
	if !ok || len(entryArr) == 0 {
		c.Status(200)
		return
	}

	entry, ok := entryArr[0].(map[string]interface{})
	if !ok {
		c.Status(200)
		return
	}

	changesArr, ok := entry["changes"].([]interface{})
	if !ok || len(changesArr) == 0 {
		c.Status(200)
		return
	}

	change, ok := changesArr[0].(map[string]interface{})
	if !ok {
		c.Status(200)
		return
	}

	value, ok := change["value"].(map[string]interface{})
	if !ok {
		c.Status(200)
		return
	}

	messages, exists := value["messages"]
	if !exists {
		c.Status(200)
		return
	}

	msgArr := messages.([]interface{})
	if len(msgArr) == 0 {
		c.Status(200)
		return
	}

	msg := msgArr[0].(map[string]interface{})
	from := msg["from"].(string)
	msgType := msg["type"].(string)

	// 📸 IMAGE MESSAGE
	if msgType == "image" {
		image := msg["image"].(map[string]interface{})
		mediaID := image["id"].(string)

		// Download image from WhatsApp
		filePath := services.DownloadWhatsAppMedia(mediaID)

		// OCR
		text, err := services.ExtractText(filePath)
		if err != nil {
			sendMessage(from, "❌ OCR failed")
			c.Status(200)
			return
		}

		amount := services.DetectAmount(text)

		if amount > 0 {
			services.AddExpense(amount, "Bill Image")
			sendMessage(from, fmt.Sprintf("🧾 ₹%.2f added", amount))
		} else {
			sendMessage(from, "❌ Amount not detected")
		}
	}

	// 📝 TEXT MESSAGE COMMANDS
	if msgType == "text" {
		textMap := msg["text"].(map[string]interface{})
		textBody := strings.ToLower(textMap["body"].(string))

		switch textBody {

		case "/expenses":
			total := services.GetTotalExpense()
			sendMessage(from, fmt.Sprintf("💰 Total Expenses: ₹%.2f", total))

		case "/statement":
			file := services.GenerateMonthlyPDF()
			sendMessage(from, fmt.Sprintf("📄 Statement generated: %s", file))

		default:
			sendMessage(from, "Send /expenses or /statement")
		}
	}

	c.Status(200)
}

func sendMessage(phone, message string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	token := os.Getenv("ACCESS_TOKEN")

	payload := fmt.Sprintf(`{
		"messaging_product": "whatsapp",
		"to": "%s",
		"type": "text",
		"text": {"body": "%s"}
	}`, phone, message)

	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	http.DefaultClient.Do(req)
}
