package handlers

import (
	"net/http"
	"twitch_chat_analysis/cmd/api/messages"
	"twitch_chat_analysis/cmd/api/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

type HandlersService struct {
	RabbitMQConn *amqp.Connection
	RedisClient  *redis.Client
	Ch           *amqp.Channel
	Q            amqp.Queue
}

// post message api function
func (h *HandlersService) PostMessage(c *gin.Context) {
	var message models.Message
	//read message
	err := c.BindJSON(&message)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Publish the message to RabbitMQ
	err = messages.PublishMessage(message, h.RabbitMQConn, h.Ch, h.Q)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// get messages api function
func (h *HandlersService) GetMessages(c *gin.Context) {
	//read param values
	sender := c.Query("sender")
	receiver := c.Query("receiver")

	// Retrieve the messages from Redis
	messages, err := messages.GetMessages(sender, receiver, h.RedisClient)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Return the messages in chronological descending order
	result := make([]models.Message, len(messages))
	for i, message := range messages {
		result[len(messages)-1-i] = message
	}

	c.JSON(http.StatusOK, result)
}
