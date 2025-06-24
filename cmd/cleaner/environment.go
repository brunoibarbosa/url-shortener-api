package main

import (
	"log"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
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
				Host:     config.MustEnv("DB_HOST"),
				User:     config.MustEnv("DB_USER"),
				Password: config.MustEnv("DB_PASSWORD"),
				Name:     config.MustEnv("DB_NAME"),
				Port:     config.MustEnvAsInt("DB_PORT"),
			},

			ExpiredURLCleanupInterval: config.MustEnvAsDuration("EXPIRED_URL_CLEANUP_INTERVAL"),
		},
	}
}
