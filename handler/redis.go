package handler

import (
	"context"
	"log"
	"twitch_chat_analysis/models"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var channelName = "report"

func PublisherMessage(message models.ResponseModel) bool {
	conn := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	err := conn.Ping(ctx).Err()
	if err != nil {
		log.Fatalln("Redis not connected. => " + err.Error())
		return false
	}

	conn.Publish(ctx, channelName, message)
	return true
}
