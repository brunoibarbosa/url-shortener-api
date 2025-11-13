package http

import (
	"fmt"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	oauth_provider "github.com/brunoibarbosa/url-shortener/internal/infra/oauth"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
	pg_session_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/session"
	pg_user_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/user"
	redis_session_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/jwt"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler/auth"
	"github.com/redis/go-redis/v9"
)

type AuthRoutesConfig struct {
	JWTSecret            string
	GoogleID             string
	GoogleSecret         string
	ListenAddress        string
	RefreshTokenDuration time.Duration
	AccessTokenDuration  time.Duration
}

func SetupAuthRoutes(r *http_router.AppRouter, pgConn *pg.Postgres, redisClient *redis.Client, config AuthRoutesConfig) {
	userRepo := pg_user_repo.NewUserRepository(pgConn)
	providerRepo := pg_user_repo.NewUserProviderRepository(pgConn)
	profileRepo := pg_user_repo.NewUserProfileRepository(pgConn)
	sessionRepo := pg_session_repo.NewSessionRepository(pgConn)
	blacklistRepo := redis_session_repo.NewBlacklistRepository(redisClient)

	provider := oauth_provider.NewGoogleOAuth(config.GoogleID, config.GoogleSecret, fmt.Sprintf("http://%s", config.ListenAddress))
	tokenService := jwt.NewTokenService(config.JWTSecret)

	// --------------------------------------------------

	registerHandler := command.NewRegisterUserHandler(
		userRepo,
		providerRepo,
		profileRepo,
	)
	registerHTTPHandler := handler.NewRegisterUserHTTPHandler(registerHandler)

	// --------------------------------------------------

	loginUserHandler := command.NewLoginUserHandler(
		providerRepo,
		sessionRepo,
		tokenService,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	loginUserHTTPHandler := handler.NewLoginUserHTTPHandler(loginUserHandler, config.RefreshTokenDuration)

	// --------------------------------------------------

	redirectGoogleHandler := command.NewRedirectGoogleHandler(provider)
	redirectGoogleHTTPHandler := handler.NewRedirectGoogleHTTPHandler(redirectGoogleHandler)

	// --------------------------------------------------

	loginGoogleHandler := command.NewLoginGoogleHandler(
		provider,
		userRepo,
		providerRepo,
		profileRepo,
		tokenService,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	loginGoogleHTTPHandler := handler.NewLoginGoogleHTTPHandler(loginGoogleHandler)

	// --------------------------------------------------

	refreshTokenHandler := command.NewRefreshTokenHandler(
		sessionRepo,
		blacklistRepo,
		tokenService,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	refreshTokenHTTPHandler := handler.NewRefreshTokenHTTPHandler(refreshTokenHandler, config.RefreshTokenDuration)

	// --------------------------------------------------

	logoutHandler := command.NewLogoutHandler(sessionRepo, blacklistRepo)
	logoutHTTPHandler := handler.NewLogoutHTTPHandler(logoutHandler)

	// --------------------------------------------------

	r.Post("/auth/register", registerHTTPHandler.Handle)
	r.Post("/auth/login", loginUserHTTPHandler.Handle)
	r.Get("/auth/google", redirectGoogleHTTPHandler.Handle)
	r.Get("/auth/google/callback", loginGoogleHTTPHandler.Handle)
	r.Post("/auth/refresh", refreshTokenHTTPHandler.Handle)
	r.Post("/auth/logout", logoutHTTPHandler.Handle)
}
