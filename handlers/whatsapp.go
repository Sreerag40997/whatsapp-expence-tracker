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

	if len(body["entry"].([]interface{})) == 0 {
		c.Status(200)
		return
	}

	entry, _ := body["entry"].([]interface{})[0].(map[string]interface{})
	change, _ := entry["changes"].([]interface{})[0].(map[string]interface{})
	value, _ := change["value"].(map[string]interface{})

	if msgs, ok := value["messages"].([]interface{}); ok && len(msgs) > 0 {
		msg := msgs[0].(map[string]interface{})
		from := msg["from"].(string)

		// 1. HANDLE BUTTON CLICKS
		if interactive, ok := msg["interactive"]; ok {
			btnID := interactive.(map[string]interface{})["button_reply"].(map[string]interface{})["id"].(string)
			handleText(from, btnID)
			c.Status(200)
			return
		}

		// 2. HANDLE MEDIA & TEXT
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

	// WELCOME / START
	if cleanText == "hi" || cleanText == "hello" || cleanText == "/start" {
		welcome := "🌟 *Premium Expense Assistant* 🌟\n━━━━━━━━━━━━━━━━━━━━\n" +
			"Hello Sir! I am your dedicated financial manager.\n\n" +
			"You can talk to me just like a friend:\n" +
			"✍️ *Text:* `Lunch 250` or `500 Fuel` \n" +
			"🎤 *Voice:* Malayalam or English audio \n" +
			"📸 *Photo:* Bill or Receipt images"
		sendButtons(from, welcome, []string{"yes_help", "no_help"}, []string{"📂 Dashboard", "❌ Close"})
		return
	}

	// MAIN MENU
	if cleanText == "yes_help" {
		menu := "📑 *Finance Control Center*\nSelect an option to manage your data:"
		sendButtons(from, menu, []string{"/statement", "/set_limit_btn", "/reset_prompt"}, []string{"📊 Statement", "🎯 Set Limit", "♻️ Reset All"})
		return
	}

	// BUDGET LIMIT
	if cleanText == "/set_limit_btn" {
		sendMessage(from, "🎯 *Target Setting*\nPlease type `limit` followed by the amount.\n\nExample: `limit 5000`")
		return
	}

	if strings.HasPrefix(cleanText, "limit ") {
		parts := strings.Fields(cleanText)
		if len(parts) == 2 {
			val, _ := strconv.ParseFloat(parts[1], 64)
			services.SetLimit(val)
			sendMessage(from, fmt.Sprintf("✅ *Budget Set!*\nYour monthly target is now *₹%.2f*.\nI will notify you if you exceed this.", val))
			sendFollowUp(from)
			return
		}
	}

	// STATEMENT
	if cleanText == "/statement" {
		summary := services.GetMonthlySummary(0, 0)
		sendMessage(from, summary)
		sendFollowUp(from)
		return
	}

	// RESET
	if cleanText == "/reset_prompt" {
		sendButtons(from, "⚠️ *System Reset*\nSir, are you sure? This will delete all records permanently.", []string{"actual_reset", "no_help"}, []string{"Confirm Reset", "Cancel"})
		return
	}

	if cleanText == "actual_reset" {
		services.ResetExpenses()
		sendMessage(from, "♻️ *Reset Successful!*\nYour account has been cleared to ₹0.00.")
		sendFollowUp(from)
		return
	}

	if cleanText == "no_help" {
		sendMessage(from, "👍 *Understood!* I'm standing by whenever you need to log an expense.")
		return
	}

	// LOG MANUAL ENTRY
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("✅ *Expense Logged!*\n━━━━━━━━━━━━━━\n🔹 *Item:* %s\n💰 *Amount:* ₹%.2f\n📅 *Date:* %s\n━━━━━━━━━━━━━━",
			note, amt, time.Now().Format("02 Jan 2006"))
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
		sendFollowUp(from)
	} else {
		sendMessage(from, "🤔 *Pardon me?*\nI couldn't quite catch that. Try: `Lunch 200` or `500 Fuel`.")
	}
}

// MEDIA PROCESSORS
func processImage(from string, image map[string]interface{}) {
	sendMessage(from, "🔍 *Analyzing your bill...*")
	path, _ := services.DownloadWhatsAppMedia(image["id"].(string))
	text, _ := services.ExtractTextFromImage(path)
	amt := services.DetectAmount(text)
	if amt > 0 {
		warn, over := services.AddExpense(amt, "Bill Image")
		res := fmt.Sprintf("📸 *Scan Complete!*\nAdded *₹%.2f* for *Bill Photo*.", amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
	} else {
		sendMessage(from, "❌ *OCR Failed:* I couldn't find a clear amount. Please type: `Item Amount`.")
	}
	sendFollowUp(from)
}

func processAudio(from string, audio map[string]interface{}) {
	sendMessage(from, "🎧 *Processing your voice note...*")
	path, _ := services.DownloadWhatsAppMedia(audio["id"].(string))
	text, _ := services.SpeechToText(path)
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("🎤 *Voice Logged!*\n*Item:* %s\n*Amount:* ₹%.2f", note, amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
	} else {
		sendMessage(from, "❌ *Audio Error:* I heard \""+text+"\" but didn't find an amount.")
	}
	sendFollowUp(from)
}

// --- HELPER METHODS ---

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
