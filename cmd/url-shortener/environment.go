package main

import (
	"log"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/joho/godotenv"
)

type Environment struct {
	URLSecret    string
	JWTSecret    string
	GoogleID     string
	GoogleSecret string

	PostgresConn pg.PostgresConnection

	RedisAddress  string
	RedisPassword string
	RedisDB       int

	URLPersistExpirationDuration time.Duration
	URLCacheExpirationDuration   time.Duration

	RefreshTokenDuration time.Duration
	AccessTokenDuration  time.Duration

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
			URLSecret:    config.MustEnv("URL_SECRET"),
			JWTSecret:    config.MustEnv("JWT_SECRET"),
			GoogleID:     config.MustEnv("GOOGLE_CLIENT_ID"),
			GoogleSecret: config.MustEnv("GOOGLE_CLIENT_SECRET"),

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

			RefreshTokenDuration: config.MustEnvAsDuration("REFRESH_TOKEN_DURATION"),
			AccessTokenDuration:  config.MustEnvAsDuration("ACCESS_TOKEN_DURATION"),

			ListenAddress: config.MustEnv("LISTEN_ADDRESS"),
		},
	}
}
