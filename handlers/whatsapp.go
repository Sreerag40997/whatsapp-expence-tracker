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

		// 2. HANDLE TEXT/MEDIA
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

	// WELCOME / HI
	if cleanText == "hi" || cleanText == "hello" || cleanText == "/start" {
		welcome := "👋 *Hello Sir! Welcome to ExpenseBot.*\n\nI track your spending via Text, Voice, or Photos."
		sendButtons(from, welcome, []string{"YES_HELP", "NO_HELP"}, []string{"Open Menu", "Dismiss"})
		return
	}

	// MAIN MENU
	if cleanText == "yes_help" {
		menu := "📑 *Main Menu*\nSelect an option below:"
		ids := []string{"/statement", "/set_limit_btn", "/reset_prompt"}
		titles := []string{"📜 Get Summary", "🎯 Set Limit", "♻️ Reset All"}
		sendButtons(from, menu, ids, titles)
		return
	}

	// SET LIMIT BUTTON CLICKED
	if cleanText == "/set_limit_btn" {
		sendMessage(from, "🎯 *Set Monthly Limit*\nPlease type `limit` followed by the amount.\n\nExample: `limit 5000`")
		return
	}

	// PROCESS "limit 5000" TEXT
	if strings.HasPrefix(cleanText, "limit ") {
		parts := strings.Fields(cleanText)
		if len(parts) == 2 {
			val, err := strconv.ParseFloat(parts[1], 64)
			if err == nil {
				services.SetLimit(val)
				sendMessage(from, fmt.Sprintf("✅ *Limit Updated!*\nYour budget is now ₹%.2f. I will warn you if you cross it.", val))
				sendFollowUp(from)
				return
			}
		}
		sendMessage(from, "❌ *Invalid Format:* Please type `limit 5000`.")
		return
	}

	// SUMMARY
	if cleanText == "/statement" {
		summary := services.GetMonthlySummary(0, 0)
		sendMessage(from, summary)
		sendFollowUp(from)
		return
	}

	// RESET LOGIC
	if cleanText == "/reset_prompt" {
		sendButtons(from, "⚠️ *Reset everything?*\nThis will delete all expenses and the limit.", []string{"ACTUAL_RESET", "no_help"}, []string{"Yes, Reset", "Cancel"})
		return
	}

	if cleanText == "actual_reset" {
		services.ResetExpenses()
		sendMessage(from, "♻️ *Reset Successful!* Your records have been cleared.")
		sendFollowUp(from)
		return
	}

	if cleanText == "no_help" {
		sendMessage(from, "👍 *Understood!* Send an expense whenever you are ready.")
		return
	}

	// MANUAL EXPENSE ENTRY (e.g., "Pizza 500")
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("✅ *Logged:* %s - ₹%.2f", note, amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
		sendFollowUp(from)
	} else {
		sendMessage(from, "🤔 *Not sure how to handle that.*\nTry: `Pizza 400` or click a button.")
	}
}

// MEDIA PROCESSORS
func processImage(from string, image map[string]interface{}) {
	sendMessage(from, "🔍 *Analyzing bill...*")
	path, _ := services.DownloadWhatsAppMedia(image["id"].(string))
	text, _ := services.ExtractTextFromImage(path)
	amt := services.DetectAmount(text)
	if amt > 0 {
		warn, over := services.AddExpense(amt, "Bill Photo")
		res := fmt.Sprintf("✅ *Photo Scanned:* ₹%.2f", amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
	} else {
		sendMessage(from, "❌ Could not find amount in photo.")
	}
	sendFollowUp(from)
}

func processAudio(from string, audio map[string]interface{}) {
	sendMessage(from, "🎧 *Processing voice...*")
	path, _ := services.DownloadWhatsAppMedia(audio["id"].(string))
	text, _ := services.SpeechToText(path)
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("🎤 *Voice Added:* %s - ₹%.2f", note, amt)
		if over {
			res += "\n\n" + warn
		}
		sendMessage(from, res)
	} else {
		sendMessage(from, "❌ Could not understand voice amount.")
	}
	sendFollowUp(from)
}

// HELPERS
func sendFollowUp(to string) {
	time.Sleep(800 * time.Millisecond)
	sendButtons(to, "🤝 Need anything else?", []string{"YES_HELP", "no_help"}, []string{"Show Menu", "I'm Done"})
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
