package main

// import (
// 	"context"
// 	"io" // Added for io.Copy
// 	"os"

// 	openai "github.com/sashabaranov/go-openai"
// )

// func main() {
// 	// Make sure your OPENAI_API_KEY is set in your environment variables
// 	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

// 	req := openai.CreateSpeechRequest{
// 		Model: openai.TTSModel1, // Recommended model for speed
// 		Input: "Food 200 രൂപ",   // Testing with Malayalam text too
// 		Voice: openai.VoiceAlloy,
// 	}

// 	// 1. Create the speech request
// 	resp, err := client.CreateSpeech(context.Background(), req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Close() // ALWAYS close the response stream

// 	// 2. Create the output file
// 	f, err := os.Create("test_voice.mp3")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	// 3. FIX: Use io.Copy to stream data from the response to the file
// 	_, err = io.Copy(f, resp)
// 	if err != nil {
// 		panic(err)
// 	}

// 	println("✅ Speech generated successfully: test_voice.mp3")
// }
