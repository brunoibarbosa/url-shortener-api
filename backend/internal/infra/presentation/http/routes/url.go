package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/config"
	memory_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/memory"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/url"
	"github.com/go-chi/chi/v5"
)

func SetupURLRoutes(r chi.Router, appConfig config.AppConfig) {
	repo := memory_repo.NewURLRepository()

	createHandler := command.NewCreateShortURLHandler(repo, appConfig.Env.SecretKey)
	getHandler := command.NewGetOriginalURLHandler(repo, appConfig.Env.SecretKey)

	createHTTPHandler := handler.NewCreateShortURLHTTPHandler(createHandler)
	redirectHTTPHandler := handler.NewRedirectHTTPHandler(getHandler)

	r.Post("/shorten", createHTTPHandler.Handle)
	r.Get("/{shortCode}", redirectHTTPHandler.Handle)
}
