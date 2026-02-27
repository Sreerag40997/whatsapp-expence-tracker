package handlers

import (
	"bot/services"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	json.NewDecoder(c.Request.Body).Decode(&body)

	entry, _ := body["entry"].([]interface{})[0].(map[string]interface{})
	change, _ := entry["changes"].([]interface{})[0].(map[string]interface{})
	value, _ := change["value"].(map[string]interface{})

	if msgs, ok := value["messages"].([]interface{}); ok && len(msgs) > 0 {
		msg := msgs[0].(map[string]interface{})
		from := msg["from"].(string)

		// HANDLE BUTTON CLICKS
		if interactive, ok := msg["interactive"]; ok {
			btnID := interactive.(map[string]interface{})["button_reply"].(map[string]interface{})["id"].(string)
			handleText(from, btnID)
			return
		}

		// HANDLE TEXT/MEDIA
		msgType := msg["type"].(string)
		switch msgType {
		case "text":
			handleText(from, msg["text"].(map[string]interface{})["body"].(string))
		case "image":
			sendMessage(from, "🔍 *Analyzing your bill image...*")
			image := msg["image"].(map[string]interface{})
			path, _ := services.DownloadWhatsAppMedia(image["id"].(string))
			text, _ := services.ExtractTextFromImage(path)
			amt := services.DetectAmount(text)
			if amt > 0 {
				warn, over := services.AddExpense(amt, "Bill Photo")
				res := fmt.Sprintf("✅ *Bill Scanned!*\nAdded ₹%.2f to your records.", amt)
				if over { res += "\n\n" + warn }
				sendMessage(from, res)
				sendFollowUp(from)
			} else {
				sendMessage(from, "❌ *OCR Failed:* I couldn't find a clear 'Total' amount. Try typing: `Lunch 200`")
				sendFollowUp(from)
			}
		case "audio":
			sendMessage(from, "🎧 *Processing your voice note...*")
			audio := msg["audio"].(map[string]interface{})
			path, _ := services.DownloadWhatsAppMedia(audio["id"].(string))
			text, _ := services.SpeechToText(path)
			note, amt, ok := services.ParseExpense(text)
			if ok {
				warn, over := services.AddExpense(amt, note)
				res := fmt.Sprintf("🎤 *Voice Note Added!*\nItem: %s\nAmount: ₹%.2f", note, amt)
				if over { res += "\n\n" + warn }
				sendMessage(from, res)
				sendFollowUp(from)
			} else {
				sendMessage(from, "❌ *Voice Error:* I heard \""+text+"\" but couldn't understand the amount. Please say something like: 'Dinner 500'")
				sendFollowUp(from)
			}
		}
	}
	c.Status(200)
}

func handleText(from, text string) {
	cleanText := strings.ToLower(strings.TrimSpace(text))

	// 1. WELCOME / HI
	if cleanText == "hi" || cleanText == "hello" || cleanText == "/start" {
		welcome := "👋 *Hello Sir! Welcome to ExpenseBot.*\n\n" +
			"I am your premium assistant for tracking daily spending. You can talk to me just like a friend!\n\n" +
			"💡 *Quick Ways to Add Expense:*\n" +
			"• ✍️ *Text:* Send `Lunch 250` or `Fuel 1000` \n" +
			"• 🎤 *Voice:* Send a voice note saying \"Food 500\"\n" +
			"• 📸 *Photo:* Send a clear picture of your bill\n\n" +
			"Would you like to see your current reports?"
		sendButtons(from, welcome, []string{"YES_HELP", "NO_HELP"}, []string{"Yes, Please", "No, Thanks"})
		return
	}

	// 2. INTERNAL FLOW: YES / NO HELP
	if cleanText == "yes_help" {
		menu := "📑 *Main Menu*\nSelect an option below to manage your finances:"
		sendButtons(from, menu, []string{"/expenses", "/statement", "/reset_prompt"}, []string{"💰 View Total", "📜 Get Bill", "♻️ Reset All"})
		return
	}

	if cleanText == "no_help" {
		sendMessage(from, "👍 *Understood!* Just send me an expense whenever you're ready. Have a productive day! 🚀")
		return
	}

	// 3. ACTUAL COMMANDS
	if cleanText == "/expenses" {
		total := services.GetTotalExpense()
		sendMessage(from, fmt.Sprintf("💵 *Current Spending Status*\n━━━━━━━━━━━━━━━━━━\n💰 *TOTAL:* ₹%.2f\n━━━━━━━━━━━━━━━━━━", total))
		sendFollowUp(from)
		return
	}

	if cleanText == "/statement" {
		sendMessage(from, services.GetMonthlySummary(0, 0))
		sendFollowUp(from)
		return
	}

	if cleanText == "/reset_prompt" {
		sendButtons(from, "⚠️ *Are you sure?* This will delete all your current month records.", []string{"ACTUAL_RESET", "no_help"}, []string{"Yes, Reset", "Cancel"})
		return
	}

	if cleanText == "actual_reset" {
		services.ResetExpenses()
		sendMessage(from, "♻️ *Reset Successful!* Your balance is now ₹0.00.")
		sendFollowUp(from)
		return
	}

	// 4. MANUAL EXPENSE ENTRY
	note, amt, ok := services.ParseExpense(text)
	if ok {
		warn, over := services.AddExpense(amt, note)
		res := fmt.Sprintf("✅ *Expense Logged!*\n━━━━━━━━━━━━━━\n🔹 *Item:* %s\n💰 *Amount:* ₹%.2f\n📅 *Date:* %s\n━━━━━━━━━━━━━━",
			note, amt, time.Now().Format("02 Jan 2006"))
		if over { res += "\n\n" + warn }
		sendMessage(from, res)
		sendFollowUp(from)
	} else {
		sendMessage(from, "🤔 *Not sure how to handle that.*\nTry: `Item Amount` (e.g., `Pizza 400`) or type *Hi* for help.")
	}
}

// HELPER: Sends the "Do you need more help?" buttons
func sendFollowUp(to string) {
	// Slight delay feels more natural
	time.Sleep(1 * time.Second)
	sendButtons(to, "🤝 *Done, Sir!* Is there anything else I can help you with today?", []string{"YES_HELP", "NO_HELP"}, []string{"Yes, show menu", "No, I'm done"})
}

// GENERIC BUTTON SENDER
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