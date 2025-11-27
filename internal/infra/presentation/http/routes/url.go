package http

import (
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/url"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/url"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/url"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type URLRoutesConfig struct {
	URLSecret                    string
	URLPersistExpirationDuration time.Duration
	URLCacheExpirationDuration   time.Duration
}

func SetupURLRoutes(r *http_router.AppRouter, pgConn *pgxpool.Pool, redisClient *redis.Client, config URLRoutesConfig) {
	repo := pg_repo.NewURLRepository(pgConn)
	cache := redis_repo.NewURLCacheRepository(redisClient)

	// --------------------------------------------------

	createHandler := command.NewCreateShortURLHandler(
		repo,
		cache,
		config.URLSecret,
		config.URLPersistExpirationDuration,
		config.URLCacheExpirationDuration,
	)
	createHTTPHandler := handler.NewCreateShortURLHTTPHandler(createHandler)

	// --------------------------------------------------

	getHandler := command.NewGetOriginalURLHandler(
		repo,
		cache,
		config.URLSecret,
		config.URLCacheExpirationDuration,
	)
	redirectHTTPHandler := handler.NewRedirectHTTPHandler(getHandler)

	// --------------------------------------------------

	r.Post("/url/shorten", createHTTPHandler.Handle)
	r.Get("/{shortCode}", redirectHTTPHandler.Handle)
}
