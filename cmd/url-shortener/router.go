package main

import (
	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	http "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http/routes"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/middleware"
	"github.com/redis/go-redis/v9"
)

func getRouter(postgres *pg.Postgres, redisClient *redis.Client, appConfig config.AppConfig) *http_router.AppRouter {
	r := http_router.NewRouter(appConfig)

	r.Use(
		middleware.LocaleMiddleware,
		middleware.RecoverMiddleware,
	)
	http.SetupURLRoutes(r, postgres, redisClient, appConfig)

	return r
}
