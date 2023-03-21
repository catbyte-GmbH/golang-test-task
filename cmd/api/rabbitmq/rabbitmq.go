package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

func InitRabbitMQ() *amqp.Connection {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://user:password@localhost:7001/")
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
