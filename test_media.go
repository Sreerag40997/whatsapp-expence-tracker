package main

// import (
// 	"bot/services"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/joho/godotenv"
// )

// func main() {
// 	// 1. Load Environment Variables (API Keys)
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	fmt.Println("🚀 --- STARTING LOCAL MEDIA TEST ---")

// 	// --- TEST 1: OCR (IMAGE) ---
// 	imagePath := "test_bill.jpeg" // Make sure this file exists in your folder
// 	if _, err := os.Stat(imagePath); err == nil {
// 		fmt.Println("\n📸 Testing OCR...")
// 		text, err := services.ExtractTextFromImage(imagePath)
// 		if err != nil {
// 			fmt.Printf("❌ OCR Error: %v\n", err)
// 		} else {
// 			amount := services.DetectAmount(text)
// 			fmt.Printf("✅ OCR Success!\n📝 Text Found: %s\n💰 Detected Amount: ₹%.2f\n", text, amount)
// 		}
// 	} else {
// 		fmt.Println("\n⚠️ Skipping OCR Test: 'test_bill.jpg' not found.")
// 	}

// 	// --- TEST 2: SPEECH (VOICE) ---
// 	voicePath := "test_voice.ogg" // Make sure this file exists in your folder
// 	if _, err := os.Stat(voicePath); err == nil {
// 		fmt.Println("\n🎧 Testing Speech-to-Text...")
// 		transcription, err := services.SpeechToText(voicePath)
// 		if err != nil {
// 			fmt.Printf("❌ Speech Error: %v\n", err)
// 		} else {
// 			note, amt, ok := services.ParseExpense(transcription)
// 			fmt.Printf("✅ Speech Success!\n🎤 Heard: \"%s\"\n🏷️ Parsed: %s - ₹%.2f (OK: %v)\n", transcription, note, amt, ok)
// 		}
// 	} else {
// 		fmt.Println("\n⚠️ Skipping Speech Test: 'test_voice.ogg' not found.")
// 	}

// 	fmt.Println("\n🏁 --- TEST COMPLETE ---")
// }
