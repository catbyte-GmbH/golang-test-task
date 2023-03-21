package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

var (
	rabbitMQConn *amqp.Connection
	redisClient  *redis.Client
	ch           *amqp.Channel
	q            amqp.Queue
)

func main() {
	// Initialize the RabbitMQ connection
	rabbitMQConn = initRabbitMQ()
	defer rabbitMQConn.Close()

	// Initialize the Redis client
	redisClient = initRedis()
	defer redisClient.Close()

	// Initialize the Gin r
	r := gin.Default()

	// Define the POST endpoint
	r.POST("/message", postMessage)

	// Define the GET endpoint
	r.GET("/message/list", GetMessages)

	var err error
	//init channel
	ch, err = rabbitMQConn.Channel()
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer ch.Close()

	// Declare the queue
	q, err = ch.QueueDeclare(
		"message",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf(err.Error())
	}

	//process messages
	go messageProcessor()

	// Start the Gin server
	if err != nil {
		log.Fatal(err)
	}
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.Run()
}

// post message api function
func postMessage(c *gin.Context) {
	var message Message
	//read message
	err := c.BindJSON(&message)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Publish the message to RabbitMQ
	err = publishMessage(message)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// get messages api function
func GetMessages(c *gin.Context) {
	//read param values
	sender := c.Query("sender")
	receiver := c.Query("receiver")

	// Retrieve the messages from Redis
	messages, err := getMessages(sender, receiver)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Return the messages in chronological descending order
	result := make([]Message, len(messages))
	for i, message := range messages {
		result[len(messages)-1-i] = message
	}

	c.JSON(http.StatusOK, result)
}

func initRabbitMQ() *amqp.Connection {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://user:password@localhost:7001/")
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func initRedis() *redis.Client {
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	//ping to make sure its alive and no issues.
	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func publishMessage(message Message) error {
	// Create a channel
	var err error
	ch, err = rabbitMQConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare the queue
	q, err = ch.QueueDeclare(
		"message",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Convert the message to JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Publish the message to the queue
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonMessage,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func getMessages(sender string, receiver string) ([]Message, error) {
	// Get the messages from Redis
	var messages []Message
	if sender == "" && receiver == "" {
		// Get all messages
		keys, err := redisClient.Keys(redisClient.Context(), "*").Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			jsonMessage, err := redisClient.Get(redisClient.Context(), key).Bytes()
			if err != nil {
				return nil, err
			}

			var message Message
			err = json.Unmarshal(jsonMessage, &message)
			if err != nil {
				return nil, err
			}

			messages = append(messages, message)
		}
	} else {
		// Get messages for a specific sender and receiver
		key := fmt.Sprintf("%s:%s", sender, receiver)
		jsonMessages, err := redisClient.LRange(redisClient.Context(), key, 0, -1).Result()
		if err != nil {
			return nil, err
		}

		for _, jsonMessage := range jsonMessages {
			var message Message
			err = json.Unmarshal([]byte(jsonMessage), &message)
			if err != nil {
				return nil, err
			}

			messages = append(messages, message)
		}
	}

	return messages, nil
}

func messageProcessor() {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// Unmarshal the message
			var message Message
			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// Save the message to Redis
			key := fmt.Sprintf("%s:%s", message.Sender, message.Receiver)
			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}
			err = redisClient.RPush(redisClient.Context(), key, jsonMessage).Err()
			if err != nil {
				log.Printf("Failed to save message to Redis: %v", err)
				continue
			}

			log.Printf("Processed message: %s", d.Body)
		}
	}()

	log.Printf("Waiting for messages...")
	<-forever
}
