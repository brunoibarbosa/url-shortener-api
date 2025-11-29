package env

import (
	"log"
	"os"
	"strconv"
	"time"
)

func GetEnv(key string) string {
	val := os.Getenv(key)
	return val
}

func MustEnv(key string) string {
	val := GetEnv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s not set", key)
	}
	return val
}

func GetEnvWithDefault(key, defaultVal string) string {
	val := GetEnv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func GetEnvAsInt(key string, defaultValue int) int {
	valStr := GetEnvWithDefault(key, strconv.Itoa(defaultValue))
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected integer, got %s", key, valStr)
	}
	return val
}

func MustEnvAsDuration(key string) time.Duration {
	valStr := MustEnv(key)

	valDuration, err := time.ParseDuration(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected time.Duration, got %s", key, valStr)
	}

	return valDuration
}

func MustEnvAsInt(key string) int {
	valStr := MustEnv(key)

	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: expected int, got %s", key, valStr)
	}

	return valInt
}
