package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/infra/app/session/query"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/session"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRoutesConfig struct {
	JWTSecret string
}

func SetupSessionRoutes(r *http_router.AppRouter, pgConn *pgxpool.Pool, config SessionRoutesConfig) {
	authMiddleware := middleware.NewAuthMiddleware(config.JWTSecret)

	listHandler := query.NewListSessionsHandler(pgConn)
	listSessiontHTTPHandler := handler.NewListSessionsHTTPHandler(listHandler)

	// --------------------------------------------------

	r.Group(
		func(r *http_router.AppRouter) {
			r.Use(authMiddleware.Handler)

			r.Get("/user/sessions", listSessiontHTTPHandler.Handle)
		},
	)
}
