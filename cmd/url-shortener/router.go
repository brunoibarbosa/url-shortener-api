package main

import (
	"github.com/brunoibarbosa/url-shortener/internal/config"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	http "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http/routes"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/middleware"
)

func getRouter(appConfig config.AppConfig) *http_router.AppRouter {
	r := http_router.NewRouter(appConfig)

	r.Use(
		middleware.LocaleMiddleware,
		middleware.RecoverMiddleware,
	)
	http.SetupURLRoutes(r, appConfig)

	return r
}
