package main

import (
	"log"

	"github.com/brunoibarbosa/url-shortener/internal/i18n"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/redis"
)

func main() {
	cfg := LoadAppConfig()

	postgres := pg.NewPostgres(cfg.Env.PostgresConn)
	defer postgres.Pool.Close()

	redisClient := redis.GetRedisClient(redis.RedisConfig{
		RedisAddress:  cfg.Env.RedisAddress,
		RedisPassword: cfg.Env.RedisPassword,
		RedisDB:       cfg.Env.RedisDB,
	})
	defer redisClient.Close()

	if err := i18n.Init(); err != nil {
		log.Fatalf("failed to initialize i18n: %v", err)
	}

	r := getRouter(postgres, redisClient, cfg)
	listenAndServe(r, cfg)
}
