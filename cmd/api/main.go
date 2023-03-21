package main

import (
	"log"
	handlers "twitch_chat_analysis/cmd/api/handlers"
	"twitch_chat_analysis/cmd/api/messages"
	"twitch_chat_analysis/cmd/api/rabbitmq"
	red "twitch_chat_analysis/cmd/api/redis"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

var (
	rabbitMQConn *amqp.Connection
	redisClient  *redis.Client
	ch           *amqp.Channel
	q            amqp.Queue
)

func init() {

}

func main() {
	// Initialize the RabbitMQ connection
	rabbitMQConn = rabbitmq.InitRabbitMQ()
	defer rabbitMQConn.Close()

	// Initialize the Redis client
	redisClient = red.InitRedis()
	defer redisClient.Close()

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
	go messages.MessageProcessor(ch, q, redisClient)

	// Start the Gin server
	if err != nil {
		log.Fatal(err)
	}

	handlersService := handlers.HandlersService{
		RabbitMQConn: rabbitMQConn,
		RedisClient:  redisClient,
		Ch:           ch,
		Q:            q,
	}
	// Initialize the Gin r
	r := gin.Default()

	// Define the POST endpoint
	r.POST("/message", handlersService.PostMessage)

	// Define the GET endpoint
	r.GET("/message/list", handlersService.GetMessages)

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, "worked")
	})

	r.Run()
}
