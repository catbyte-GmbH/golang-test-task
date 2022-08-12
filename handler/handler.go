package handler

import (
	"log"
	"net/http"
	"twitch_chat_analysis/models"

	"github.com/gin-gonic/gin"
)

func Message(c *gin.Context) {
	var q models.ResponseModel
	// Gelen JSON içeriğinin quote yapısına eşlenip eşlenemediği kontrol ediliyor
	if err := c.ShouldBindJSON(&q); err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": err}) // Eğer JSON içeriğinde sıkıntı varsa hata mesajı ile birlikte geriye HTTP 400 Bad Request dönüyoruz
		return
	}
	ok := RabbitQueuePush(q)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Success": q})

}

func MessageList(c *gin.Context) {
	MessageProcessor()	
	
	c.JSON(http.StatusOK, gin.H{"Success": "test"})
}