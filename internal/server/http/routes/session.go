package http_routes

import (
	"github.com/brunoibarbosa/url-shortener/internal/infra/app/session/query"
	"github.com/brunoibarbosa/url-shortener/internal/server/http"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler/session"
	http_middleware "github.com/brunoibarbosa/url-shortener/internal/server/http/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRoutesConfig struct {
	JWTSecret string
}

func NewSessionRoutes(r *http.AppRouter, pgConn *pgxpool.Pool, config SessionRoutesConfig) {
	authMiddleware := http_middleware.NewAuthMiddleware(config.JWTSecret)

	listHandler := query.NewListSessionsHandler(pgConn)
	listSessiontHTTPHandler := http_handler.NewListSessionsHTTPHandler(listHandler)

	// --------------------------------------------------

	r.Group(
		func(r *http.AppRouter) {
			r.Use(authMiddleware.Handler)

			r.Get("/user/sessions", listSessiontHTTPHandler.Handle)
		},
	)
}
