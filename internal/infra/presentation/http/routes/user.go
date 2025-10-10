package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/app/user/command"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/user"
)

type UserRoutesConfig struct {
	JWTSecret string
}

func SetupUserRoutes(r *http_router.AppRouter, pgConn *pg.Postgres, config UserRoutesConfig) {
	repo := pg_repo.NewUserRepository(pgConn)

	registerHandler := command.NewRegisterUserHandler(repo)
	loginHandler := command.NewLoginUserHandler(repo, config.JWTSecret)

	registerHTTPHandler := handler.NewRegisterUserHTTPHandler(registerHandler)
	loginHTTPHandler := handler.NewLoginUserHTTPHandler(loginHandler)

	r.Post("/register", registerHTTPHandler.Handle)
	r.Post("/login", loginHTTPHandler.Handle)
}
