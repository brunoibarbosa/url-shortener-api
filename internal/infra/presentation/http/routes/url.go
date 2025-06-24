package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/url"
	"github.com/redis/go-redis/v9"
)

func SetupURLRoutes(r *http_router.AppRouter, pgConn *pg.Postgres, redisClient *redis.Client, appConfig config.AppConfig) {
	repo := pg_repo.NewURLRepository(pgConn)
	cache := redis_repo.NewURLCacheRepository(redisClient)

	createHandler := command.NewCreateShortURLHandler(repo, cache, appConfig.Env.SecretKey, appConfig.Env.ExpireDuration)
	getHandler := command.NewGetOriginalURLHandler(repo, cache, appConfig.Env.SecretKey, appConfig.Env.ExpireDuration)

	createHTTPHandler := handler.NewCreateShortURLHTTPHandler(createHandler)
	redirectHTTPHandler := handler.NewRedirectHTTPHandler(getHandler)

	r.Post("/shorten", createHTTPHandler.Handle)
	r.Get("/{shortCode}", redirectHTTPHandler.Handle)
}
