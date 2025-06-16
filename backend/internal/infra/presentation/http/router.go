package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/config"
	routes "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http/routes"
	"github.com/go-chi/chi/v5"
)

func NewRouter(appConfig config.AppConfig) *chi.Mux {
	r := chi.NewRouter()

	routes.SetupURLRoutes(r, appConfig)

	return r
}
