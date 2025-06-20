package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/config"
	redis "github.com/brunoibarbosa/url-shortener/internal/infra/database"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/url"
)

func SetupURLRoutes(r *http_router.AppRouter, appConfig config.AppConfig) {
	redisClient := redis.GetRedisClient(appConfig)
	repo := redis_repo.NewURLRepository(redisClient)

	createHandler := command.NewCreateShortURLHandler(repo, appConfig.Env.SecretKey, appConfig.Env.ExpireDuration)
	getHandler := command.NewGetOriginalURLHandler(repo, appConfig.Env.SecretKey)

	createHTTPHandler := handler.NewCreateShortURLHTTPHandler(createHandler)
	redirectHTTPHandler := handler.NewRedirectHTTPHandler(getHandler)

	r.Post("/shorten", createHTTPHandler.Handle)
	r.Get("/{shortCode}", redirectHTTPHandler.Handle)
}
