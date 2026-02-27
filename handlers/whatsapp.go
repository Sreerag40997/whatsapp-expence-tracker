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

func VerifyWebhook(c *gin.Context) {
	if c.Query("hub.verify_token") == os.Getenv("VERIFY_TOKEN") {
		c.String(200, c.Query("hub.challenge"))
		return
	}
	c.Status(403)
}

func ReceiveMessage(c *gin.Context) {
	var body map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&body)

	entry, _ := body["entry"].([]interface{})[0].(map[string]interface{})
	change, _ := entry["changes"].([]interface{})[0].(map[string]interface{})
	value, _ := change["value"].(map[string]interface{})

	// 1. Detect Interactive Button Clicks
	if msgs, ok := value["messages"].([]interface{}); ok && len(msgs) > 0 {
		msg := msgs[0].(map[string]interface{})
		from := msg["from"].(string)

		if interactive, ok := msg["interactive"]; ok {
			btnID := interactive.(map[string]interface{})["button_reply"].(map[string]interface{})["id"].(string)
			handleText(from, btnID)
			return
		}

		// 2. Handle standard message types
		msgType := msg["type"].(string)
		switch msgType {
		case "text":
			handleText(from, msg["text"].(map[string]interface{})["body"].(string))
		case "image":
			image := msg["image"].(map[string]interface{})
			path, _ := services.DownloadWhatsAppMedia(image["id"].(string))
			text, _ := services.ExtractTextFromImage(path)
			amt := services.DetectAmount(text)
			if amt > 0 {
				warn, over := services.AddExpense(amt, "Bill OCR")
				reply := fmt.Sprintf("✅ *Bill Detected!*\nAmount: ₹%.2f", amt)
				if over {
					reply += "\n\n" + warn
				}
				sendMessage(from, reply)
			} else {
				sendMessage(from, "❌ Could not find Total on bill. Send text: 'Item Amount'")
			}
		case "audio":
			audio := msg["audio"].(map[string]interface{})
			path, _ := services.DownloadWhatsAppMedia(audio["id"].(string))
			text, _ := services.SpeechToText(path)
			note, amt, ok := services.ParseExpense(text)
			if ok {
				warn, over := services.AddExpense(amt, note)
				reply := fmt.Sprintf("🎤 *Recorded:* %s (₹%.2f)", note, amt)
				if over {
					reply += "\n\n" + warn
				}
				sendMessage(from, reply)
			} else {
				sendMessage(from, "🎤 Heard: \""+text+"\"\n❌ Format failed. Try: 'Lunch 200'")
			}
		}
	}
	c.Status(200)
}

func handleText(from, text string) {
	text = strings.ToLower(strings.TrimSpace(text))

	// Main Commands
	if text == "hi" || text == "/start" || text == "menu" {
		sendButtons(from, "Hello Sir! 📊\nHow can I help you manage your money today?", []string{"/expenses", "/statement", "/reset-expense"})
		return
	}

	if strings.HasPrefix(text, "/statement") {
		// Handles: /statement OR /statement/05/2024
		parts := strings.Split(text, "/")
		m, y := 0, 0
		if len(parts) == 4 {
			m, _ = strconv.Atoi(parts[2])
			y, _ = strconv.Atoi(parts[3])
		}
		summary := services.GetMonthlySummary(m, y)
		sendMessage(from, summary)
		return
	}

	if text == "/expenses" {
		total := services.GetTotalExpense()
		sendMessage(from, fmt.Sprintf("💰 *Current Total:* ₹%.2f", total))
		return
	}

	if text == "/reset-expense" {
		services.ResetExpenses()
		sendMessage(from, "♻️ *Expenses Cleared!* Total set to ₹0.00")
		return
	}

	if strings.HasPrefix(text, "/setlimit=") {
		val, _ := strconv.ParseFloat(strings.Split(text, "=")[1], 64)
		services.SetLimit(val)
		sendMessage(from, fmt.Sprintf("✅ *Limit Set:* ₹%.2f", val))
		return
	}

	// Parsing manual entry like "Food 500"
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		reply := fmt.Sprintf("✅ *Added:* %s (₹%.2f)", note, amt)
		if over {
			reply += "\n\n" + warn
		}
		sendMessage(from, reply)
	} else {
		sendMessage(from, "❓ Unknown format. Try 'Lunch 200' or use /start for menu.")
	}
}

func sendMessage(to, text string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	payload := map[string]interface{}{
		"messaging_product": "whatsapp", "to": to, "type": "text", "text": map[string]string{"body": text},
	}
	sendRequest(url, payload)
}

func sendButtons(to, bodyText string, buttonIDs []string) {
	url := "https://graph.facebook.com/v18.0/" + os.Getenv("PHONE_NUMBER_ID") + "/messages"
	btns := []map[string]interface{}{}
	for _, id := range buttonIDs {
		btns = append(btns, map[string]interface{}{
			"type": "reply", "reply": map[string]string{"id": id, "title": id},
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

func sendRequest(url string, payload interface{}) {
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(data)))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}
