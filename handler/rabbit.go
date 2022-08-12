package handler

import (
	"encoding/json"
	"log"
	"twitch_chat_analysis/models"

	"github.com/streadway/amqp"
)

func RabbitQueuePush(message models.ResponseModel) bool {
	rabbit, err := amqp.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		log.Fatalln(err)
	}

	ch, err := rabbit.Channel()
	if err != nil {
		log.Fatalln(err)
		return false
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalln(err)
		return false
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		})
	if err != nil {
		log.Fatalln(err)
		return false
	}
	return true

}
