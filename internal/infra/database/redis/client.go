package redis

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

type RedisConfig struct {
	RedisAddress  string
	RedisPassword string
	RedisDB       int
}

func GetRedisClient(redisConfig RedisConfig) *redis.Client {
	once.Do(func() {
		log.Printf("Connecting to Redis at %s (DB: %d)...", redisConfig.RedisAddress, redisConfig.RedisDB)

		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConfig.RedisAddress,
			Password: redisConfig.RedisPassword,
			DB:       redisConfig.RedisDB,
		})

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Failed to connect to Redis: %v", err)
		} else {
			log.Printf("Successfully connected to Redis at %s", redisConfig.RedisAddress)
		}
	})
	return redisClient
}
