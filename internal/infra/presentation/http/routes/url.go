package http

import (
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/url"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/url"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/url"
	"github.com/redis/go-redis/v9"
)

type URLRoutesConfig struct {
	URLSecret                    string
	URLPersistExpirationDuration time.Duration
	URLCacheExpirationDuration   time.Duration
}

func SetupURLRoutes(r *http_router.AppRouter, pgConn *pg.Postgres, redisClient *redis.Client, config URLRoutesConfig) {
	repo := pg_repo.NewURLRepository(pgConn)
	cache := redis_repo.NewURLCacheRepository(redisClient)

	createHandler := command.NewCreateShortURLHandler(repo, cache, config.URLSecret, config.URLPersistExpirationDuration, config.URLCacheExpirationDuration)
	getHandler := command.NewGetOriginalURLHandler(repo, cache, config.URLSecret, config.URLCacheExpirationDuration)

	createHTTPHandler := handler.NewCreateShortURLHTTPHandler(createHandler)
	redirectHTTPHandler := handler.NewRedirectHTTPHandler(getHandler)

	r.Post("/url/shorten", createHTTPHandler.Handle)
	r.Get("/{shortCode}", redirectHTTPHandler.Handle)
}
