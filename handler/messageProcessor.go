package handler

import (
	"encoding/json"
	"log"

	"twitch_chat_analysis/models"

	"github.com/streadway/amqp"
)



func MessageProcessor() bool {
	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		log.Fatalln(err)
		return false
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalln(err)
		return false
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"test_queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalln(err)
		return false
	}

	var a models.ResponseModel

	for d := range msgs {
		json.Unmarshal(d.Body, &a)
		ok := PublisherMessage(a)
		if !ok {
			log.Fatalln(err)
			return false
		}
	}
	return true

}
