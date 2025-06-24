package redis

import (
	"sync"

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
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConfig.RedisAddress,
			Password: redisConfig.RedisPassword,
			DB:       redisConfig.RedisDB,
		})
	})
	return redisClient
}
