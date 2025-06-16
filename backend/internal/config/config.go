package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Environment struct {
	SecretKey     string
	RedisAddress  string
	RedisPassword string
	RedisDB       int
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
			SecretKey:     mustEnv("SECRET_KEY"),
			RedisAddress:  mustEnv("REDIS_ADDRESS"),
			RedisPassword: getEnvWithDefault("REDIS_PASSWORD", ""),
			RedisDB:       getEnvAsInt("REDIS_DB", 0),
		},
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s not set", key)
	}
	return val
}

func getEnvWithDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvAsInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected integer, got %s", key, valStr)
	}
	return val
}
