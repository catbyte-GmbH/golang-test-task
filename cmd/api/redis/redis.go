package redis

import (
	"log"

	"github.com/go-redis/redis/v8"
)

func InitRedis() *redis.Client {
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
