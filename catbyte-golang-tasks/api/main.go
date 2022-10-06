package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"twitch_chat_analysis/jsonhandling"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func rabbitmqConn(message jsonhandling.Message) {

	conn, err0 := amqp.Dial("amqp://user:password@localhost:7001/")
	jsonhandling.HandleError(err0, "Cannot connect to AMQP")
	defer conn.Close()

	ch, err1 := conn.Channel()
	jsonhandling.HandleError(err1, "Cannot create a amqpChannel")
	defer ch.Close()

	q, err2 := ch.QueueDeclare(
		"messagesQueue",
		false,
		false,
		false,
		false,
		nil,
	)
	jsonhandling.HandleError(err2, "Could not declare `messagesQueue` queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := jsonhandling.JsonMarshal(message)

	err3 := ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	jsonhandling.HandleError(err3, "Error publishing message")

	log.Println("Message published to the queue successfully :)")

}

func postHttpRequestJson(context *gin.Context) {
	var newMessage jsonhandling.Message

	if err := context.BindJSON(&newMessage); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"response": "bad request"})
	} else {
		now := time.Now()
		newMessage.CreatedAt, newMessage.UpdatedAt = now, now
		rabbitmqConn(newMessage)
		context.IndentedJSON(http.StatusOK, newMessage)
	}
}

func main() {

	router := gin.Default()

	router.POST("/message", postHttpRequestJson)

	err := router.Run("localhost:8082")
	if err != nil {
		return
	}

}
