package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type PostgresConnection struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     int
}

type Environment struct {
	SecretKey string

	PostgresConn PostgresConnection

	RedisAddress  string
	RedisPassword string
	RedisDB       int

	ExpireDuration time.Duration

	ListenAddress string
}

type AppConfig struct {
	Env Environment
}

func Load() AppConfig {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	return AppConfig{
		Env: Environment{
			SecretKey: mustEnv("SECRET_KEY"),

			PostgresConn: PostgresConnection{
				Host:     mustEnv("DB_HOST"),
				User:     mustEnv("DB_USER"),
				Password: mustEnv("DB_PASSWORD"),
				Name:     mustEnv("DB_NAME"),
				Port:     mustEnvAsInt("DB_PORT"),
			},

			RedisAddress:  mustEnv("REDIS_ADDRESS"),
			RedisPassword: getEnvWithDefault("REDIS_PASSWORD", ""),
			RedisDB:       getEnvAsInt("REDIS_DB", 0),

			ExpireDuration: mustEnvAsDuration("EXPIRE_DURATION"),

			ListenAddress: mustEnv("LISTEN_ADDRESS"),
		},
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	return val
}

func mustEnv(key string) string {
	val := getEnv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s not set", key)
	}
	return val
}

func getEnvWithDefault(key, defaultVal string) string {
	val := getEnv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvAsInt(key string, defaultValue int) int {
	valStr := getEnvWithDefault(key, strconv.Itoa(defaultValue))
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected integer, got %s", key, valStr)
	}
	return val
}

func mustEnvAsDuration(key string) time.Duration {
	valStr := mustEnv(key)

	valDuration, err := time.ParseDuration(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected time.Duration, got %s", key, valStr)
	}

	return valDuration
}

func mustEnvAsInt(key string) int {
	valStr := mustEnv(key)

	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected int, got %s", key, valStr)
	}

	return valInt
}
