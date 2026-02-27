package main

import (
	"bot/handlers"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	os.Mkdir("tmp", os.ModePerm)

	r := gin.Default()
	r.Static("/public", "./public")

	r.GET("/webhook", handlers.VerifyWebhook)
	r.POST("/webhook", handlers.ReceiveMessage)

	r.Run(":8080")
}
