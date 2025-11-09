package http

import (
	"fmt"

	"github.com/brunoibarbosa/url-shortener/internal/app/user/command"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	oauth_provider "github.com/brunoibarbosa/url-shortener/internal/infra/oauth"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_session_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/session"
	pg_user_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/jwt"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/user"
)

type UserRoutesConfig struct {
	JWTSecret     string
	GoogleID      string
	GoogleSecret  string
	ListenAddress string
}

func SetupUserRoutes(r *http_router.AppRouter, pgConn *pg.Postgres, config UserRoutesConfig) {
	userRepo := pg_user_repo.NewUserRepository(pgConn)
	providerRepo := pg_user_repo.NewUserProviderRepository(pgConn)
	profileRepo := pg_user_repo.NewUserProfileRepository(pgConn)
	sessionRepo := pg_session_repo.NewSessionRepository(pgConn)

	provider := oauth_provider.NewGoogleOAuth(config.GoogleID, config.GoogleSecret, fmt.Sprintf("http://%s", config.ListenAddress))
	tokenService := jwt.NewTokenService(config.JWTSecret)

	registerHandler := command.NewRegisterUserHandler(userRepo, providerRepo, profileRepo)
	registerHTTPHandler := handler.NewRegisterUserHTTPHandler(registerHandler)

	loginUserHandler := command.NewLoginUserHandler(providerRepo, sessionRepo, tokenService)
	loginUserHTTPHandler := handler.NewLoginUserHTTPHandler(loginUserHandler)

	redirectGoogleHandler := command.NewRedirectGoogleHandler(provider)
	redirectGoogleHTTPHandler := handler.NewRedirectGoogleHTTPHandler(redirectGoogleHandler)

	loginGoogleHandler := command.NewLoginGoogleHandler(provider, userRepo, providerRepo, profileRepo, tokenService)
	loginGoogleHTTPHandler := handler.NewLoginGoogleHTTPHandler(loginGoogleHandler)

	r.Post("/register", registerHTTPHandler.Handle)
	r.Post("/login", loginUserHTTPHandler.Handle)

	r.Get("/auth/google", redirectGoogleHTTPHandler.Handle)
	r.Get("/auth/google/callback", loginGoogleHTTPHandler.Handle)
}
