package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/brunoibarbosa/url-shortener/internal/i18n"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/redis"
	"github.com/brunoibarbosa/url-shortener/internal/server/http"
	http_middleware "github.com/brunoibarbosa/url-shortener/internal/server/http/middleware"
	http_routes "github.com/brunoibarbosa/url-shortener/internal/server/http/routes"
)

func main() {
	log.Println("Starting URL Shortener API...")

	// Config
	log.Println("Loading application configuration...")
	cfg := LoadAppConfig()

	// Database
	log.Println("Initializing database connections...")
	postgres := pg.NewPostgres(cfg.Env.PostgresConn)
	defer postgres.Pool.Close()

	redisClient := redis.GetRedisClient(redis.RedisConfig{
		RedisAddress:  cfg.Env.RedisAddress,
		RedisPassword: cfg.Env.RedisPassword,
		RedisDB:       cfg.Env.RedisDB,
	})
	defer redisClient.Close()

	// Translation
	log.Println("Initializing i18n translations...")
	if err := i18n.Init(); err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}
	log.Println("i18n translations initialized successfully")

	// Router
	log.Println("Setting up HTTP router and routes...")
	router := http.NewRouter()
	router.Use(
		http_middleware.LocaleMiddleware,
		http_middleware.RecoverMiddleware,
	)
	http_routes.NewURLRoutes(router, postgres.Pool, redisClient, http_routes.URLRoutesConfig{
		URLSecret:                    cfg.Env.URLSecret,
		URLPersistExpirationDuration: cfg.Env.URLPersistExpirationDuration,
		URLCacheExpirationDuration:   cfg.Env.URLCacheExpirationDuration,
	})
	http_routes.NewAuthRoutes(router, postgres.Pool, redisClient, http_routes.AuthRoutesConfig{
		JWTSecret:            cfg.Env.JWTSecret,
		GoogleID:             cfg.Env.GoogleID,
		GoogleSecret:         cfg.Env.GoogleSecret,
		ListenAddress:        cfg.Env.ListenAddress,
		RefreshTokenDuration: cfg.Env.RefreshTokenDuration,
		AccessTokenDuration:  cfg.Env.AccessTokenDuration,
	})
	http_routes.NewSessionRoutes(router, postgres.Pool, http_routes.SessionRoutesConfig{
		JWTSecret: cfg.Env.JWTSecret,
	})

	// Swagger - usa caminho absoluto para evitar problemas com diretório de trabalho
	swaggerSpecPath := filepath.Join(getProjectRoot(), "docs", "openapi", "openapi.yaml")
	http_routes.NewSwaggerRoutes(router, http_routes.SwaggerRoutesConfig{
		SpecPath: swaggerSpecPath,
	})
	log.Println("Routes configured successfully")

	// Server
	log.Printf("Starting HTTP server on %s...", cfg.Env.ListenAddress)
	server := http.NewServer(cfg.Env.ListenAddress, router)
	log.Printf("Server is ready and listening on %s", cfg.Env.ListenAddress)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// getProjectRoot retorna o caminho absoluto para a raiz do projeto
func getProjectRoot() string {
	// Obtém o diretório atual do arquivo main.go
	// Este arquivo está em cmd/url-shortener, então precisamos subir 2 níveis
	executable, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Se estiver executando de cmd/url-shortener, volta 2 níveis
	if filepath.Base(executable) == "url-shortener" && filepath.Base(filepath.Dir(executable)) == "cmd" {
		return filepath.Join(executable, "..", "..")
	}

	// Caso contrário, assume que já está na raiz
	return executable
}
