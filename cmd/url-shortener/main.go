package main

import (
	"log"

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
	log.Println("Routes configured successfully")

	// Server
	log.Printf("Starting HTTP server on %s...", cfg.Env.ListenAddress)
	server := http.NewServer(cfg.Env.ListenAddress, router)
	log.Printf("Server is ready and listening on %s", cfg.Env.ListenAddress)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
