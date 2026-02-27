package handlers

import (
	"bot/services"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// VerifyWebhook handles the initial Meta/Facebook verification
func VerifyWebhook(c *gin.Context) {
	verifyToken := os.Getenv("VERIFY_TOKEN")
	if c.Query("hub.mode") == "subscribe" && c.Query("hub.verify_token") == verifyToken {
		c.String(200, c.Query("hub.challenge"))
		return
	}
	c.Status(403)
}

// ReceiveMessage is the main webhook handler for incoming WhatsApp messages
func ReceiveMessage(c *gin.Context) {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.Status(200)
		return
	}

	// Navigate through the WhatsApp JSON structure
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
			warn, isOver := services.AddExpense(amt, "Bill Image")
			reply := fmt.Sprintf("🧾 *Bill Detected!*\n\n*Amount:* ₹%.2f\n*Status:* Added to records.", amt)
			if isOver {
				reply += "\n\n" + warn
			}
			sendMessage(from, reply)
		} else {
			sendMessage(from, "❌ Could not detect a total amount on this bill. Please ensure the 'Total' is clearly visible.")
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
			sendMessage(from, "❌ Voice recognition failed. Ensure API key is valid.")
			break
		}

		note, amt, ok := services.ParseExpense(text)
		if ok {
			warn, isOver := services.AddExpense(amt, note)
			reply := fmt.Sprintf("🎤 *Voice Record Added!*\n\n*Item:* %s\n*Amount:* ₹%.2f", note, amt)
			if isOver {
				reply += "\n\n" + warn
			}
			sendMessage(from, reply)
		} else {
			sendMessage(from, fmt.Sprintf("🎤 I heard: \"%s\"\n\n❌ Format not recognized. Try saying: 'Dinner 500'", text))
		}
	}

	c.Status(200)
}

func handleText(from, text string) {
	cleanText := strings.ToLower(strings.TrimSpace(text))

	// 1. GREETINGS
	if cleanText == "hi" || cleanText == "hello" || cleanText == "hlo" {
		greeting := "Hello Sir! 👋\n\n" +
			"I am your *Personal Expense Tracker Bot*.\n\n" +
			"*Commands:*\n" +
			"💰 /expenses — View total\n" +
			"📊 /setlimit=700 — Set budget\n" +
			"♻️ /reset-expense — Clear all\n" +
			"📄 /statement — Get PDF\n\n" +
			"Or just send: 'Coffee 100' or a Voice/Photo!"
		sendMessage(from, greeting)
		return
	}

	// 2. SET LIMIT COMMAND (e.g., /setlimit=1000)
	if strings.HasPrefix(cleanText, "/setlimit=") {
		parts := strings.Split(cleanText, "=")
		if len(parts) == 2 {
			limit, err := strconv.ParseFloat(parts[1], 64)
			if err == nil {
				services.SetLimit(limit)
				sendMessage(from, fmt.Sprintf("✅ *Limit Set!*\nYour new budget limit is ₹%.2f.", limit))
				return
			}
		}
		sendMessage(from, "❌ Invalid format. Use: /setlimit=500")
		return
	}

	// 3. RESET EXPENSES COMMAND
	if cleanText == "/reset-expense" {
		services.ResetExpenses()
		sendMessage(from, "♻️ *Reset Successful!*\nYour expenses have been set to ₹0.00.")
		return
	}

	// 4. TOTAL EXPENSES COMMAND
	if cleanText == "/expenses" {
		total := services.GetTotalExpense()
		sendMessage(from, fmt.Sprintf("💰 *Current Total:* ₹%.2f", total))
		return
	}

	// 5. PDF STATEMENT COMMAND
	if cleanText == "/statement" {
		sendMessage(from, "⏳ Generating your monthly statement...")
		file := services.GenerateMonthlyPDF()
		sendDocument(from, file)
		return
	}

	// 6. MANUAL EXPENSE ENTRY (e.g., "Food 200")
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, isOver := services.AddExpense(amt, note)
		reply := fmt.Sprintf("✅ *Expense Added!*\n\n*Item:* %s\n*Amount:* ₹%.2f", note, amt)
		if isOver {
			reply += "\n\n" + warn
		}
		sendMessage(from, reply)
	} else {
		sendMessage(from, "❌ *Invalid Format*\n\nPlease send like: 'Dinner 500'\nOr use /expenses to see your total.")
	}
}

// sendMessage sends a standard text message back to the user
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

// sendDocument sends the generated PDF statement
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

// sendRequest executes the HTTP POST to Meta Graph API
func sendRequest(url string, payload interface{}) {
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(data)))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
}
