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
	"github.com/jackc/pgx/v5/pgxpool"
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

func SetupAuthRoutes(r *http_router.AppRouter, pgConn *pgxpool.Pool, redisClient *redis.Client, config AuthRoutesConfig) {
	txManager := pg.NewTxManager(pgConn)

	userRepo := pg_user_repo.NewUserRepository(pgConn)
	providerRepo := pg_user_repo.NewUserProviderRepository(pgConn)
	profileRepo := pg_user_repo.NewUserProfileRepository(pgConn)
	sessionRepo := pg_session_repo.NewSessionRepository(pgConn)
	blacklistRepo := redis_session_repo.NewBlacklistRepository(redisClient)

	provider := oauth_provider.NewGoogleOAuth(config.GoogleID, config.GoogleSecret, fmt.Sprintf("http://%s", config.ListenAddress))
	tokenService := jwt.NewTokenService(config.JWTSecret)

	// --------------------------------------------------

	registerHandler := command.NewRegisterUserHandler(
		txManager,
		userRepo,
		providerRepo,
		profileRepo,
	)
	registerHTTPHandler := handler.NewRegisterUserHTTPHandler(registerHandler)

	// --------------------------------------------------

	loginUserHandler := command.NewLoginUserHandler(
		txManager,
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
		txManager,
		provider,
		userRepo,
		providerRepo,
		profileRepo,
		sessionRepo,
		tokenService,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	loginGoogleHTTPHandler := handler.NewLoginGoogleHTTPHandler(loginGoogleHandler, config.RefreshTokenDuration)

	// --------------------------------------------------

	refreshTokenHandler := command.NewRefreshTokenHandler(
		txManager,
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
