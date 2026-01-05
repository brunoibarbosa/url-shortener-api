package http_routes

import (
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/container"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/url"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/crypto"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/shortcode"
	"github.com/brunoibarbosa/url-shortener/internal/server/http"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler/url"
	http_middleware "github.com/brunoibarbosa/url-shortener/internal/server/http/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type URLRoutesConfig struct {
	JWTSecret                    string
	URLSecret                    string
	URLPersistExpirationDuration time.Duration
	URLCacheExpirationDuration   time.Duration
}

func NewURLRoutes(r *http.AppRouter, pgConn *pgxpool.Pool, redisClient *redis.Client, config URLRoutesConfig) {
	optionalAuth := http_middleware.NewOptionalAuthMiddleware(config.JWTSecret)

	deps := container.URLFactoryDependencies{
		PersistRepo:               pg_repo.NewURLRepository(pgConn),
		CacheRepo:                 redis_repo.NewURLCacheRepository(redisClient),
		Encrypter:                 crypto.NewURLEncrypter(config.URLSecret),
		ShortCodeGenerator:        shortcode.NewRandomShortCodeGenerator(),
		PersistExpirationDuration: config.URLPersistExpirationDuration,
		CacheExpirationDuration:   config.URLCacheExpirationDuration,
	}

	f := container.NewURLHandlerFactory(deps)

	createHTTPHandler := http_handler.NewCreateShortURLHTTPHandler(f.CreateShortURLHandler())
	redirectHTTPHandler := http_handler.NewRedirectHTTPHandler(f.GetOriginalURLHandler())

	r.Group(func(r *http.AppRouter) {
		r.Use(optionalAuth.Handler)
		r.Post("/url/shorten", createHTTPHandler.Handle)
	})
	r.Get("/r/{shortCode}", redirectHTTPHandler.Handle)
}
