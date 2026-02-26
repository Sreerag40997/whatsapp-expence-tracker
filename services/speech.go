package services

import "fmt"


func SpeechToText(audioPath string) (string, error) {
	
	return "Dinner 500", nil 
}

func HandleVoiceExpense(audioPath string) (float64, string, error) {
	text, err := SpeechToText(audioPath)
	if err != nil {
		return 0, "", fmt.Errorf("speech to text failed: %v", err)
	}

	note, amount, ok := ParseExpense(text)
	if !ok {
		return 0, "", fmt.Errorf("could not parse expense from text: %s", text)
	}

	AddExpense(amount, note)

	return amount, note, nil
}