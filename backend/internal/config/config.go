package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	SecretKey string
}

var Env *Environment

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	Env = &Environment{
		SecretKey: mustEnv("SECRET_KEY"),
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s not set", key)
	}
	return val
}
