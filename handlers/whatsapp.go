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

	// 📸 IMAGE MESSAGE
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
			services.AppendRow(fmt.Sprintf("Bill Image ₹%.2f", amount))
			sendMessage(from, fmt.Sprintf("🧾 Expense Added: ₹%.2f", amount))
		} else {
			sendMessage(from, "❌ Amount not detected")
		}
	}

	// 📝 TEXT MESSAGE
	if msgType == "text" {
		textBody := strings.ToLower(msg["text"].(map[string]interface{})["body"].(string))

		reply := handleUserText(from, textBody)
		sendMessage(from, reply)
	}

	c.Status(200)
}

func handleUserText(user, text string) string {

	// 👋 Greeting
	if text == "hi" || text == "hello" || text == "hlo" {
		return "👋 Welcome to Expense Tracker Bot\n\n" +
			"Send:\n" +
			"📝 Lunch 200\n" +
			"📄 /statement\n" +
			"💰 /expenses"
	}

	// 📄 Statement
	if text == "/statement" {
		file := services.GenerateMonthlyPDF()
		return fmt.Sprintf("📄 Statement generated: %s", file)
	}

	// 💰 Total Expenses
	if text == "/expenses" {
		total := services.GetTotalExpense()
		return fmt.Sprintf("💰 Total Expenses: ₹%.2f", total)
	}

	// 🧾 Detect normal expense text: "Lunch 200"
	ok, title, amount := services.ParseExpenseText(text)
	if ok {
		services.AppendRow(fmt.Sprintf("%s ₹%.2f", title, amount))
		return fmt.Sprintf("✅ Added: %s ₹%.2f", title, amount)
	}

	return "❌ Invalid format.\nSend like: Lunch 200"
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
