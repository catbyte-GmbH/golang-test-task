package service

import (
	hand "twitch_chat_analysis/handler"

	"github.com/gin-gonic/gin"
)

func ServiceRun() {
	r := gin.Default()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.POST("/message", hand.Message)
	r.GET("/message/list", hand.MessageList)
	r.Run()
}
