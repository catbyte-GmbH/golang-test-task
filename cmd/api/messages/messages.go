package messages

import (
	"encoding/json"
	"fmt"
	"log"
	"twitch_chat_analysis/cmd/api/models"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

func MessageProcessor(ch *amqp.Channel, q amqp.Queue, redisClient *redis.Client) {
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
			var message models.Message
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

func PublishMessage(message models.Message, rabbitMQConn *amqp.Connection, ch *amqp.Channel, q amqp.Queue) error {
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

func GetMessages(sender string, receiver string, redisClient *redis.Client) ([]models.Message, error) {
	// Get the messages from Redis
	var messages []models.Message
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

			var message models.Message
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
			var message models.Message
			err = json.Unmarshal([]byte(jsonMessage), &message)
			if err != nil {
				return nil, err
			}

			messages = append(messages, message)
		}
	}

	return messages, nil
}
