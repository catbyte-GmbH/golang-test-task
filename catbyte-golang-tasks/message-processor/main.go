package main

import (
	"log"
	"time"

	"twitch_chat_analysis/jsonhandling"
	db "twitch_chat_analysis/radis-db"

	amqp "github.com/rabbitmq/amqp091-go"
)

var messages = []jsonhandling.Message{
	{
		Sender:      "Bob",
		Receiver:    "Alice",
		MessageBody: "Goodbye Alice.",
		CreatedAt:   time.Now().Add(24 * time.Hour),
		UpdatedAt:   time.Now().Add(24 * time.Hour),
	},
	{
		Sender:      "Bob",
		Receiver:    "Alice",
		MessageBody: "Hey Alice, What a great profile you have.",
	},
	{
		Sender:      "Bob",
		Receiver:    "Alice",
		MessageBody: "Hey Alice, I'm glad to connect with you.",
		CreatedAt:   time.Now(),
	},
	{
		Sender:      "Bob",
		Receiver:    "Alice",
		MessageBody: "Hey Alice, That's a great thing to keep in touch with you.",
		CreatedAt:   time.Now().Add(12 * time.Hour),
		UpdatedAt:   time.Now().Add(12 * time.Hour),
	},
	{
		Sender:      "Alice",
		Receiver:    "Bob",
		MessageBody: "Hey Bob, I'm glad to connect with you.",
		CreatedAt:   time.Now(),
	},
}

func redisConn(msgs []jsonhandling.Message) {
	database, err := db.NewClient("localhost:6379")
	db.Must(err)

	savedMessage := jsonhandling.RadisJsonMarshal(msgs)
	db.Set(database, "messages", savedMessage, 0)

	log.Printf("Message saved successfully to redis database:\n%s", savedMessage)
}

func rabbitmqConsumer() {

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

	msgs, err3 := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	jsonhandling.HandleError(err3, "Could not register consumer")

	forever := make(chan string)

	redisConn(messages)
	go func() {
		for d := range msgs {
			msg := jsonhandling.Message{}
			messages = append(messages, jsonhandling.JsonUnmarshal(d.Body, msg))

			redisConn(messages)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func main() {
	rabbitmqConsumer()
}
