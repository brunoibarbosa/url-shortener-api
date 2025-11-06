package http

import (
	"fmt"

	"github.com/brunoibarbosa/url-shortener/internal/app/user/command"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	oauth_provider "github.com/brunoibarbosa/url-shortener/internal/infra/oauth"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/jwt"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/user"
	"golang.org/x/crypto/bcrypt"
)

type UserRoutesConfig struct {
	JWTSecret     string
	GoogleID      string
	GoogleSecret  string
	ListenAddress string
}

func SetupUserRoutes(r *http_router.AppRouter, pgConn *pg.Postgres, config UserRoutesConfig) {
	userRepo := pg_repo.NewUserRepository(pgConn)
	providerRepo := pg_repo.NewUserProviderRepository(pgConn)
	profileRepo := pg_repo.NewUserProfileRepository(pgConn)

	provider := oauth_provider.NewGoogleOAuth(config.GoogleID, config.GoogleSecret, fmt.Sprintf("http://%s", config.ListenAddress))
	tokenService := jwt.NewTokenService(config.JWTSecret)

	registerHandler := command.NewRegisterUserHandler(userRepo, providerRepo, profileRepo, hashPassword)
	registerHTTPHandler := handler.NewRegisterUserHTTPHandler(registerHandler)

	loginUserHandler := command.NewLoginUserHandler(providerRepo, tokenService, checkPasswordHash)
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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
