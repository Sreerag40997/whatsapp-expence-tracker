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

	entryArr, ok := body["entry"].([]interface{})
	if !ok || len(entryArr) == 0 {
		c.Status(200)
		return
	}

	entry := entryArr[0].(map[string]interface{})
	changesArr, ok := entry["changes"].([]interface{})
	if !ok || len(changesArr) == 0 {
		c.Status(200)
		return
	}

	change := changesArr[0].(map[string]interface{})
	value := change["value"].(map[string]interface{})

	msgArr, ok := value["messages"].([]interface{})
	if !ok || len(msgArr) == 0 {
		c.Status(200)
		return
	}

	msg := msgArr[0].(map[string]interface{})
	from := msg["from"].(string)
	msgType := msg["type"].(string)

	switch msgType {

	case "text":
		text := msg["text"].(map[string]interface{})["body"].(string)
		handleText(from, text)

	case "image":
		image := msg["image"].(map[string]interface{})
		path, err := services.DownloadWhatsAppMedia(image["id"].(string))
		if err != nil {
			sendMessage(from, "❌ Failed to download image")
			break
		}

		ocrText, err := services.ExtractTextFromImage(path)
		if err != nil {
			sendMessage(from, "❌ OCR failed on image")
			break
		}

		amt := services.DetectAmount(ocrText)
		if amt > 0 {
			services.AddExpense(amt, "Bill OCR") 
			sendMessage(from, fmt.Sprintf("🧾 Added from Bill: ₹%.2f", amt))
		} else {
			sendMessage(from, "❌ Could not detect amount on bill. Try clear image.")
		}

	case "audio":
		audio := msg["audio"].(map[string]interface{})
		path, err := services.DownloadWhatsAppMedia(audio["id"].(string))
		if err != nil {
			sendMessage(from, "❌ Failed to download audio")
			break
		}

		text, err := services.SpeechToText(path)
		if err != nil {
			sendMessage(from, "❌ Voice recognition failed")
			break
		}

		note, amt, ok := services.ParseExpense(text)
		if ok {
			services.AddExpense(amt, note)
			sendMessage(from, fmt.Sprintf("🎤 Added: %s ₹%.2f", note, amt))
		} else {
			sendMessage(from, "❌ Speak like: Food 200")
		}
	}

	c.Status(200)
}

func handleText(from, text string) {
	cleanText := strings.ToLower(strings.TrimSpace(text))

	if cleanText == "hi" || cleanText == "hello" || cleanText == "hlo" {
		greeting := "Hello Sir! 👋\n\n" +
			"I am your *Personal Expense Tracker Bot*, designed to help you track your spending and keep your finances organized. 📊\n\n" +
			"*How to record an expense:*\n" +
			"📝 *Text:* Type 'Lunch 200' or 'Petrol 500'\n" +
			"🎤 *Voice:* Send a voice note like \"Food 300\"\n" +
			"📸 *Photo:* Send a clear picture of your bill\n\n" +
			"*Reports & Commands:*\n" +
			"💰 /expenses — View total expenses.\n" +
			"📄 /statement — Get your monthly PDF statement.\n\n" +
			"Simply send your first expense to get started!"

		sendMessage(from, greeting)
		return
	}

	if cleanText == "/expenses" {
		total := services.GetTotalExpense()
		sendMessage(from, fmt.Sprintf("💰 *Sir, your current total expense is:* ₹%.2f", total))
		return
	}

	if cleanText == "/statement" {
		sendMessage(from, "Generating your monthly statement, please wait... ⏳")
		file := services.GenerateMonthlyPDF()
		sendDocument(from, file)
		return
	}

	note, amt, ok := services.ParseExpense(text)
	if ok {
		services.AddExpense(amt, note)
		sendMessage(from, fmt.Sprintf("✅ *Expense Added, Sir!*\n\n*Item:* %s\n*Amount:* ₹%.2f", note, amt))
	} else {
		sendMessage(from, "❌ *Invalid Format, Sir.*\n\nPlease send like: 'Dinner 500'\nOr use /expenses to see your total.")
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
			"filename": "Monthly_Statement.pdf",
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