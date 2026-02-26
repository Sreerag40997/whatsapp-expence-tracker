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

	entry := entryArr[0].(map[string]interface{})
	change := entry["changes"].([]interface{})[0].(map[string]interface{})
	value := change["value"].(map[string]interface{})

	messages, exists := value["messages"]
	if !exists {
		c.Status(200)
		return
	}

	msg := messages.([]interface{})[0].(map[string]interface{})
	from := msg["from"].(string)
	msgType := msg["type"].(string)

	// 📸 IMAGE MESSAGE (Bill OCR)
	if msgType == "image" {
		image := msg["image"].(map[string]interface{})
		mediaID := image["id"].(string)

		filePath := services.DownloadWhatsAppMedia(mediaID)
		text, err := services.ExtractText(filePath)
		if err != nil {
			sendMessage(from, "❌ OCR failed")
			c.Status(200)
			return
		}

		amount := services.DetectAmount(text)
		if amount > 0 {
			services.AddExpense(amount, "Bill Image")
			services.AppendRow(fmt.Sprintf("Bill Image ₹%.2f", amount))
			sendMessage(from, fmt.Sprintf("🧾 Expense Added: ₹%.2f", amount))
		} else {
			sendMessage(from, "❌ Amount not detected")
		}
	}

	// 📝 TEXT MESSAGE
	if msgType == "text" {
		textBody := strings.ToLower(strings.TrimSpace(msg["text"].(map[string]interface{})["body"].(string)))

		replyType, reply := handleUserText(from, textBody)

		if replyType == "text" {
			sendMessage(from, reply)
		} else if replyType == "pdf" {
			sendDocument(from, reply) // reply contains file path
		}
	}

	c.Status(200)
}

func handleUserText(user, text string) (string, string) {

	// 👋 Greeting
	if text == "hi" || text == "hello" || text == "hlo" {
		return "text", "👋 Welcome to Expense Tracker Bot\n\nSend:\n📝 Lunch 200\n📸 Bill image\n💰 /expenses\n📄 /statement"
	}

	// 📄 Statement PDF
	if text == "/statement" {
		filePath := services.GenerateMonthlyPDF()
		return "pdf", filePath
	}

	// 💰 Total Expenses
	if text == "/expenses" {
		total := services.GetTotalExpense()
		return "text", fmt.Sprintf("💰 Total Expenses: ₹%.2f", total)
	}

	// 🧾 Parse: "Lunch 200"
	ok, title, amount := services.ParseExpenseText(text)
	if ok {
		services.AddExpense(amount, title)
		services.AppendRow(fmt.Sprintf("%s ₹%.2f", title, amount))
		return "text", fmt.Sprintf("✅ Added: %s ₹%.2f", title, amount)
	}

	return "text", "❌ Invalid format.\nSend like: Lunch 200"
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

func sendDocument(phone, filePath string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	token := os.Getenv("ACCESS_TOKEN")

	fileURL := os.Getenv("BASE_URL") + "/public/" + filePath

	payload := fmt.Sprintf(`{
		"messaging_product": "whatsapp",
		"to": "%s",
		"type": "document",
		"document": {
			"link": "%s",
			"filename": "%s"
		}
	}`, phone, fileURL, filePath)

	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	http.DefaultClient.Do(req)
}
