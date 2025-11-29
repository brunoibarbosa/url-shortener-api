package http_routes

import (
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/url"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/crypto"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/shortcode"
	"github.com/brunoibarbosa/url-shortener/internal/server/http"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler/url"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type URLRoutesConfig struct {
	URLSecret                    string
	URLPersistExpirationDuration time.Duration
	URLCacheExpirationDuration   time.Duration
}

func NewURLRoutes(r *http.AppRouter, pgConn *pgxpool.Pool, redisClient *redis.Client, config URLRoutesConfig) {
	repo := pg_repo.NewURLRepository(pgConn)
	cache := redis_repo.NewURLCacheRepository(redisClient)
	encrypter := crypto.NewURLEncrypter(config.URLSecret)
	shortCodeGenerator := shortcode.NewRandomShortCodeGenerator()

	// --------------------------------------------------

	createHandler := command.NewCreateShortURLHandler(
		repo,
		cache,
		encrypter,
		shortCodeGenerator,
		config.URLPersistExpirationDuration,
		config.URLCacheExpirationDuration,
	)
	createHTTPHandler := http_handler.NewCreateShortURLHTTPHandler(createHandler)

	// --------------------------------------------------

	getHandler := command.NewGetOriginalURLHandler(
		repo,
		cache,
		encrypter,
		config.URLCacheExpirationDuration,
	)
	redirectHTTPHandler := http_handler.NewRedirectHTTPHandler(getHandler)

	// --------------------------------------------------

	r.Post("/url/shorten", createHTTPHandler.Handle)
	r.Get("/r/{shortCode}", redirectHTTPHandler.Handle)
}
