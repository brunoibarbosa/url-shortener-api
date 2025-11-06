package main

import (
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	http "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http/routes"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/middleware"
	"github.com/redis/go-redis/v9"
)

func getRouter(postgres *pg.Postgres, redisClient *redis.Client, appConfig AppConfig) *http_router.AppRouter {
	r := http_router.NewRouter()

	r.Use(
		middleware.LocaleMiddleware,
		middleware.RecoverMiddleware,
	)
	http.SetupURLRoutes(r, postgres, redisClient, http.URLRoutesConfig{
		URLSecret:                    appConfig.Env.URLSecret,
		URLPersistExpirationDuration: appConfig.Env.URLPersistExpirationDuration,
		URLCacheExpirationDuration:   appConfig.Env.URLCacheExpirationDuration,
	})
	http.SetupUserRoutes(r, postgres, http.UserRoutesConfig{
		JWTSecret:     appConfig.Env.JWTSecret,
		GoogleID:      appConfig.Env.GoogleID,
		GoogleSecret:  appConfig.Env.GoogleSecret,
		ListenAddress: appConfig.Env.ListenAddress,
	})

	return r
}
