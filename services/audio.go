package services

import (
	"os/exec"
	"strings"
)

func ConvertToMP3(input string) string {
	output := strings.Replace(input, ".ogg", ".mp3", 1)

	cmd := exec.Command("ffmpeg", "-i", input, output)
	cmd.Run()

	return output
}
