package main

import (
	"log"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/pkg/env"
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
			URLSecret:    env.MustEnv("URL_SECRET"),
			JWTSecret:    env.MustEnv("JWT_SECRET"),
			GoogleID:     env.MustEnv("GOOGLE_CLIENT_ID"),
			GoogleSecret: env.MustEnv("GOOGLE_CLIENT_SECRET"),

			PostgresConn: pg.PostgresConnection{
				Host:     env.MustEnv("DB_HOST"),
				User:     env.MustEnv("DB_USER"),
				Password: env.MustEnv("DB_PASSWORD"),
				Name:     env.MustEnv("DB_NAME"),
				Port:     env.MustEnvAsInt("DB_PORT"),
			},

			RedisAddress:  env.MustEnv("REDIS_ADDRESS"),
			RedisPassword: env.GetEnvWithDefault("REDIS_PASSWORD", ""),
			RedisDB:       env.GetEnvAsInt("REDIS_DB", 0),

			URLPersistExpirationDuration: env.MustEnvAsDuration("URL_PERSIST_EXPIRATION_DURATION"),
			URLCacheExpirationDuration:   env.MustEnvAsDuration("URL_CACHE_EXPIRATION_DURATION"),

			RefreshTokenDuration: env.MustEnvAsDuration("REFRESH_TOKEN_DURATION"),
			AccessTokenDuration:  env.MustEnvAsDuration("ACCESS_TOKEN_DURATION"),

			ListenAddress: env.MustEnv("LISTEN_ADDRESS"),
		},
	}
}
