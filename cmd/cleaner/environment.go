package main

import (
	"log"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/pkg/env"
	"github.com/joho/godotenv"
)

type Environment struct {
	PostgresConn pg.PostgresConnection

	ExpiredURLCleanupInterval time.Duration
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
			PostgresConn: pg.PostgresConnection{
				Host:     env.MustEnv("DB_HOST"),
				User:     env.MustEnv("DB_USER"),
				Password: env.MustEnv("DB_PASSWORD"),
				Name:     env.MustEnv("DB_NAME"),
				Port:     env.MustEnvAsInt("DB_PORT"),
			},

			ExpiredURLCleanupInterval: env.MustEnvAsDuration("EXPIRED_URL_CLEANUP_INTERVAL"),
		},
	}
}
