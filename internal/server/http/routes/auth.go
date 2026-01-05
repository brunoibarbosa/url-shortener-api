package http_routes

import (
	"fmt"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/container"
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
	deps := container.AuthFactoryDependencies{
		TxManager:            pg.NewTxManager(pgConn),
		UserRepo:             pg_user_repo.NewUserRepository(pgConn),
		ProviderRepo:         pg_user_repo.NewUserProviderRepository(pgConn),
		ProfileRepo:          pg_user_repo.NewUserProfileRepository(pgConn),
		SessionRepo:          pg_session_repo.NewSessionRepository(pgConn),
		BlacklistRepo:        redis_session_repo.NewBlacklistRepository(redisClient),
		StateService:         redis_session_repo.NewStateRepository(redisClient),
		OAuthProvider:        oauth_provider.NewGoogleOAuth(config.GoogleID, config.GoogleSecret, fmt.Sprintf("http://%s", config.ListenAddress)),
		TokenService:         jwt.NewTokenService(config.JWTSecret),
		PasswordEncrypter:    crypto.NewUserPasswordEncrypter(bcrypt.DefaultCost),
		SessionEncrypter:     crypto.NewSessionEncrypter(),
		RefreshTokenDuration: config.RefreshTokenDuration,
		AccessTokenDuration:  config.AccessTokenDuration,
	}

	f := container.NewAuthHandlerFactory(deps)

	registerHTTPHandler := http_handler.NewRegisterUserHTTPHandler(f.RegisterUserHandler())
	loginUserHTTPHandler := http_handler.NewLoginUserHTTPHandler(f.LoginUserHandler(), f.RefreshTokenDuration())
	redirectGoogleHTTPHandler := http_handler.NewRedirectGoogleHTTPHandler(f.RedirectGoogleHandler())
	loginGoogleHTTPHandler := http_handler.NewLoginGoogleHTTPHandler(f.LoginGoogleHandler(), f.RefreshTokenDuration())
	refreshTokenHTTPHandler := http_handler.NewRefreshTokenHTTPHandler(f.RefreshTokenHandler(), f.RefreshTokenDuration())
	logoutHTTPHandler := http_handler.NewLogoutHTTPHandler(f.LogoutHandler())

	r.Post("/auth/register", registerHTTPHandler.Handle)
	r.Post("/auth/login", loginUserHTTPHandler.Handle)
	r.Get("/auth/google", redirectGoogleHTTPHandler.Handle)
	r.Get("/auth/google/callback", loginGoogleHTTPHandler.Handle)
	r.Post("/auth/refresh", refreshTokenHTTPHandler.Handle)
	r.Post("/auth/logout", logoutHTTPHandler.Handle)
}
