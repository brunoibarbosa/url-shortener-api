package main

import (
	"log"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/joho/godotenv"
)

type Environment struct {
	SecretKey string

	PostgresConn pg.PostgresConnection

	RedisAddress  string
	RedisPassword string
	RedisDB       int

	URLPersistExpirationDuration time.Duration
	URLCacheExpirationDuration   time.Duration

	ListenAddress string
}

type AppConfig struct {
	Env Environment
}

func LoadAppConfig() AppConfig {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	return AppConfig{
		Env: Environment{
			SecretKey: config.MustEnv("SECRET_KEY"),

			PostgresConn: pg.PostgresConnection{
				Host:     config.MustEnv("DB_HOST"),
				User:     config.MustEnv("DB_USER"),
				Password: config.MustEnv("DB_PASSWORD"),
				Name:     config.MustEnv("DB_NAME"),
				Port:     config.MustEnvAsInt("DB_PORT"),
			},

			RedisAddress:  config.MustEnv("REDIS_ADDRESS"),
			RedisPassword: config.GetEnvWithDefault("REDIS_PASSWORD", ""),
			RedisDB:       config.GetEnvAsInt("REDIS_DB", 0),

			URLPersistExpirationDuration: config.MustEnvAsDuration("URL_PERSIST_EXPIRATION_DURATION"),
			URLCacheExpirationDuration:   config.MustEnvAsDuration("URL_CACHE_EXPIRATION_DURATION"),

			ListenAddress: config.MustEnv("LISTEN_ADDRESS"),
		},
	}
}
