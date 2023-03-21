package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

func main() {
	r := gin.Default()

	r.POST("/message", pushToRabitMQ)
	r.GET("/message/list", handleGetMessageList)

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.Run()
}

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func pushToRabitMQ(c *gin.Context) {
	var message Message
	err := c.BindJSON(&message)
	if err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	// Validate message fields
	if message.Sender == "" || message.Receiver == "" || message.Message == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing fields"})
		return
	}

	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://user:password@localhost:7001/")
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error opening channel: %v", err)
	}
	defer ch.Close()

	// Declare a queue
	q, err := ch.QueueDeclare(
		"message_queue", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		log.Fatalf("Error declaring queue: %v", err)
	}

	// Convert message to JSON
	jsonBody, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message to JSON: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	// Publish message to queue
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		},
	)
	if err != nil {
		log.Fatalf("Error publishing message to queue: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func handleGetMessageList(c *gin.Context) {
	sender := c.Query("sender")
	receiver := c.Query("receiver")

	// Connect to Redis server
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Get messages from Redis
	key := fmt.Sprintf("%s:%s", sender, receiver)
	messages, err := rdb.ZRevRangeWithScores(context.Background(), key, 0, -1).Result()
	if err != nil {
		log.Printf("Error getting messages from Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Parse messages and create response
	var response []Message
	for _, message := range messages {
		response = append(response, Message{
			Sender:   sender,
			Receiver: receiver,
			Message:  message.Member.(string),
		})
	}

	c.JSON(http.StatusOK, response)
}
