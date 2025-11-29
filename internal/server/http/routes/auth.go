package http_routes

import (
	"fmt"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	oauth_provider "github.com/brunoibarbosa/url-shortener/internal/infra/oauth"
	pg_session_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/session"
	pg_user_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/user"
	redis_session_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/crypto"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/jwt"
	"github.com/brunoibarbosa/url-shortener/internal/server/http"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthRoutesConfig struct {
	JWTSecret            string
	GoogleID             string
	GoogleSecret         string
	ListenAddress        string
	RefreshTokenDuration time.Duration
	AccessTokenDuration  time.Duration
}

func NewAuthRoutes(r *http.AppRouter, pgConn *pgxpool.Pool, redisClient *redis.Client, config AuthRoutesConfig) {
	txManager := pg.NewTxManager(pgConn)

	userRepo := pg_user_repo.NewUserRepository(pgConn)
	providerRepo := pg_user_repo.NewUserProviderRepository(pgConn)
	profileRepo := pg_user_repo.NewUserProfileRepository(pgConn)
	sessionRepo := pg_session_repo.NewSessionRepository(pgConn)
	blacklistRepo := redis_session_repo.NewBlacklistRepository(redisClient)
	stateRepo := redis_session_repo.NewStateRepository(redisClient)

	provider := oauth_provider.NewGoogleOAuth(config.GoogleID, config.GoogleSecret, fmt.Sprintf("http://%s", config.ListenAddress))
	tokenService := jwt.NewTokenService(config.JWTSecret)
	passwordEncrypter := crypto.NewUserPasswordEncrypter(bcrypt.DefaultCost)
	sessionEncrypter := crypto.NewSessionEncrypter()

	// --------------------------------------------------

	registerHandler := command.NewRegisterUserHandler(
		txManager,
		userRepo,
		providerRepo,
		profileRepo,
		passwordEncrypter,
	)
	registerHTTPHandler := http_handler.NewRegisterUserHTTPHandler(registerHandler)

	// --------------------------------------------------

	loginUserHandler := command.NewLoginUserHandler(
		txManager,
		providerRepo,
		sessionRepo,
		tokenService,
		passwordEncrypter,
		sessionEncrypter,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	loginUserHTTPHandler := http_handler.NewLoginUserHTTPHandler(loginUserHandler, config.RefreshTokenDuration)

	// --------------------------------------------------

	redirectGoogleHandler := command.NewRedirectGoogleHandler(provider, stateRepo)
	redirectGoogleHTTPHandler := http_handler.NewRedirectGoogleHTTPHandler(redirectGoogleHandler)

	// --------------------------------------------------

	loginGoogleHandler := command.NewLoginGoogleHandler(
		txManager,
		provider,
		userRepo,
		providerRepo,
		profileRepo,
		sessionRepo,
		tokenService,
		sessionEncrypter,
		stateRepo,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	loginGoogleHTTPHandler := http_handler.NewLoginGoogleHTTPHandler(loginGoogleHandler, config.RefreshTokenDuration)

	// --------------------------------------------------

	refreshTokenHandler := command.NewRefreshTokenHandler(
		txManager,
		sessionRepo,
		blacklistRepo,
		tokenService,
		sessionEncrypter,
		config.RefreshTokenDuration,
		config.AccessTokenDuration,
	)
	refreshTokenHTTPHandler := http_handler.NewRefreshTokenHTTPHandler(refreshTokenHandler, config.RefreshTokenDuration)

	// --------------------------------------------------

	logoutHandler := command.NewLogoutHandler(sessionRepo, blacklistRepo, sessionEncrypter)
	logoutHTTPHandler := http_handler.NewLogoutHTTPHandler(logoutHandler)

	// --------------------------------------------------

	r.Post("/auth/register", registerHTTPHandler.Handle)
	r.Post("/auth/login", loginUserHTTPHandler.Handle)
	r.Get("/auth/google", redirectGoogleHTTPHandler.Handle)
	r.Get("/auth/google/callback", loginGoogleHTTPHandler.Handle)
	r.Post("/auth/refresh", refreshTokenHTTPHandler.Handle)
	r.Post("/auth/logout", logoutHTTPHandler.Handle)
}
