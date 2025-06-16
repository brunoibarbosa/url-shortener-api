package redis

import (
	"sync"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func GetRedisClient(appConfig config.AppConfig) *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     appConfig.Env.RedisAddress,
			Password: appConfig.Env.RedisPassword,
			DB:       appConfig.Env.RedisDB,
		})
	})
	return redisClient
}
