package handlers

import (
	"bot/services"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func VerifyWebhook(c *gin.Context) {
	if c.Query("hub.verify_token") == os.Getenv("VERIFY_TOKEN") {
		c.String(200, c.Query("hub.challenge"))
		return
	}
	c.Status(403)
}

func ReceiveMessage(c *gin.Context) {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.Status(400)
		return
	}

	entry, _ := body["entry"].([]interface{})[0].(map[string]interface{})
	change, _ := entry["changes"].([]interface{})[0].(map[string]interface{})
	value, _ := change["value"].(map[string]interface{})

	if msgs, ok := value["messages"].([]interface{}); ok && len(msgs) > 0 {
		msg := msgs[0].(map[string]interface{})
		from := msg["from"].(string)

		// 1. BUTTON CLICKS
		if interactive, ok := msg["interactive"]; ok {
			btnID := interactive.(map[string]interface{})["button_reply"].(map[string]interface{})["id"].(string)
			handleText(from, btnID)
			return
		}

		// 2. TEXT/MEDIA ROUTING
		msgType, _ := msg["type"].(string)
		switch msgType {
		case "text":
			handleText(from, msg["text"].(map[string]interface{})["body"].(string))
		case "image":
			processImage(from, msg["image"].(map[string]interface{}))
		case "audio":
			processAudio(from, msg["audio"].(map[string]interface{}))
		}
	}
	c.Status(200)
}

func handleText(from, text string) {
	cleanText := strings.ToLower(strings.TrimSpace(text))

	// WELCOME SCREEN
	if cleanText == "hi" || cleanText == "hello" || cleanText == "/start" {
		welcome := "🌟 *Premium Expense Assistant* 🌟\n━━━━━━━━━━━━━━━━━━━━\n" +
			"Hello Sir! I am ready to manage your finances.\n\n" +
			"✍️ *Text:* `Lunch 250` \n" +
			"🎤 *Voice:* Malayalam/English notes\n" +
			"📸 *Photo:* Bill or Receipts\n" +
			"━━━━━━━━━━━━━━━━━━━━"
		sendButtons(from, welcome, []string{"yes_help", "no_help"}, []string{"📂 Dashboard", "❌ Close"})
		return
	}

	if cleanText == "yes_help" {
		menu := "📑 *Financial Control Center*\nSelect an option below:"
		sendButtons(from, menu, []string{"/statement", "/set_limit_btn", "/reset_prompt"}, []string{"📊 Statement", "🎯 Set Limit", "♻️ Reset All"})
		return
	}

	if cleanText == "/set_limit_btn" {
		sendMessage(from, "🎯 *Target Setting*\nPlease type `limit` followed by the amount.\n\nExample: `limit 5000`")
		return
	}

	if strings.HasPrefix(cleanText, "limit ") {
		parts := strings.Fields(cleanText)
		if len(parts) == 2 {
			val, _ := strconv.ParseFloat(parts[1], 64)
			services.SetLimit(val)
			sendMessage(from, fmt.Sprintf("✅ *Budget Set!*\nMonthly target is now *₹%.2f*.", val))
			sendFollowUp(from)
			return
		}
	}

	if cleanText == "/statement" {
		sendMessage(from, services.GetMonthlySummary(0, 0))
		sendFollowUp(from)
		return
	}

	if cleanText == "/reset_prompt" {
		sendButtons(from, "⚠️ *Confirm Reset?*\nThis will permanently delete all logs.", []string{"actual_reset", "no_help"}, []string{"Confirm Reset", "Cancel"})
		return
	}

	if cleanText == "actual_reset" {
		services.ResetExpenses()
		sendMessage(from, "♻️ *System Reset Successful!*")
		sendFollowUp(from)
		return
	}

	// MANUAL PARSING
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("✅ *Expense Logged!*\n━━━━━━━━━━━━━━\n🔹 *Item:* %s\n💰 *Amount:* ₹%.2f\n━━━━━━━━━━━━━━", note, amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
		sendFollowUp(from)
	} else if cleanText != "no_help" {
		sendMessage(from, "🤔 *Pardon me?* I didn't catch that. Try: `Coffee 50`.")
	}
}

// MEDIA HANDLERS
func processImage(from string, image map[string]interface{}) {
	sendMessage(from, "🔍 *Analyzing bill...*")
	path, err := services.DownloadWhatsAppMedia(image["id"].(string))
	if err != nil {
		sendMessage(from, "❌ *Download Error:* I couldn't retrieve the image.")
		return
	}
	text, _ := services.ExtractTextFromImage(path)
	amt := services.DetectAmount(text)
	if amt > 0 {
		warn, over := services.AddExpense(amt, "Bill Photo")
		res := fmt.Sprintf("📸 *Scan Complete!*\nAdded *₹%.2f* for *Bill Photo*.", amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
	} else {
		sendMessage(from, "❌ *OCR Error:* Amount not detected. Please type manually.")
	}
	sendFollowUp(from)
}

func processAudio(from string, audio map[string]interface{}) {
	sendMessage(from, "🎧 *Processing voice...*")
	path, _ := services.DownloadWhatsAppMedia(audio["id"].(string))
	text, err := services.SpeechToText(path)

	// Handling the Quota Error
	if err != nil && err.Error() == "OPENAI_QUOTA_EXCEEDED" {
		sendMessage(from, "⚠️ *Voice Service Busy:* My transcription limit is finished. Please type your expense manually for now!")
		return
	}

	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("🎤 *Voice Added!*\n*Item:* %s\n*Amount:* ₹%.2f", note, amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
	} else {
		sendMessage(from, "❌ *Error:* I heard \""+text+"\" but no amount found.")
	}
	sendFollowUp(from)
}

// HELPER METHODS
func sendFollowUp(to string) {
	time.Sleep(1 * time.Second)
	sendButtons(to, "🤝 Need anything else, Sir?", []string{"yes_help", "no_help"}, []string{"Main Menu", "I'm Done"})
}

func sendButtons(to, bodyText string, ids []string, titles []string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	btns := []map[string]interface{}{}
	for i := range ids {
		btns = append(btns, map[string]interface{}{
			"type": "reply", "reply": map[string]string{"id": ids[i], "title": titles[i]},
		})
	}
	payload := map[string]interface{}{
		"messaging_product": "whatsapp", "to": to, "type": "interactive",
		"interactive": map[string]interface{}{
			"type": "button", "body": map[string]string{"text": bodyText},
			"action": map[string]interface{}{"buttons": btns},
		},
	}
	sendRequest(url, payload)
}

func sendMessage(to, text string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	payload := map[string]interface{}{
		"messaging_product": "whatsapp", "to": to, "type": "text", "text": map[string]string{"body": text},
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
