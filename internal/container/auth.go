package container

import (
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	bd_domain "github.com/brunoibarbosa/url-shortener/internal/domain/bd"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

type AuthHandlerFactory struct {
	txManager            bd_domain.TransactionManager
	userRepo             user_domain.UserRepository
	providerRepo         user_domain.UserProviderRepository
	profileRepo          user_domain.UserProfileRepository
	sessionRepo          session_domain.SessionRepository
	blacklistRepo        session_domain.BlacklistRepository
	stateService         session_domain.StateService
	oauthProvider        session_domain.OAuthProvider
	tokenService         session_domain.TokenService
	passwordEncrypter    user_domain.UserPasswordEncrypter
	sessionEncrypter     session_domain.SessionEncrypter
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration

	registerHandler       *command.RegisterUserHandler
	loginUserHandler      *command.LoginUserHandler
	redirectGoogleHandler *command.RedirectGoogleHandler
	loginGoogleHandler    *command.LoginGoogleHandler
	refreshTokenHandler   *command.RefreshTokenHandler
	logoutHandler         *command.LogoutHandler
}

type AuthFactoryDependencies struct {
	TxManager            bd_domain.TransactionManager
	UserRepo             user_domain.UserRepository
	ProviderRepo         user_domain.UserProviderRepository
	ProfileRepo          user_domain.UserProfileRepository
	SessionRepo          session_domain.SessionRepository
	BlacklistRepo        session_domain.BlacklistRepository
	StateService         session_domain.StateService
	OAuthProvider        session_domain.OAuthProvider
	TokenService         session_domain.TokenService
	PasswordEncrypter    user_domain.UserPasswordEncrypter
	SessionEncrypter     session_domain.SessionEncrypter
	RefreshTokenDuration time.Duration
	AccessTokenDuration  time.Duration
}

func NewAuthHandlerFactory(deps AuthFactoryDependencies) *AuthHandlerFactory {
	return &AuthHandlerFactory{
		txManager:            deps.TxManager,
		userRepo:             deps.UserRepo,
		providerRepo:         deps.ProviderRepo,
		profileRepo:          deps.ProfileRepo,
		sessionRepo:          deps.SessionRepo,
		blacklistRepo:        deps.BlacklistRepo,
		stateService:         deps.StateService,
		oauthProvider:        deps.OAuthProvider,
		tokenService:         deps.TokenService,
		passwordEncrypter:    deps.PasswordEncrypter,
		sessionEncrypter:     deps.SessionEncrypter,
		refreshTokenDuration: deps.RefreshTokenDuration,
		accessTokenDuration:  deps.AccessTokenDuration,
	}
}

func (f *AuthHandlerFactory) RegisterUserHandler() *command.RegisterUserHandler {
	if f.registerHandler == nil {
		f.registerHandler = command.NewRegisterUserHandler(
			f.txManager,
			f.userRepo,
			f.providerRepo,
			f.profileRepo,
			f.passwordEncrypter,
		)
	}
	return f.registerHandler
}

func (f *AuthHandlerFactory) LoginUserHandler() *command.LoginUserHandler {
	if f.loginUserHandler == nil {
		f.loginUserHandler = command.NewLoginUserHandler(
			f.txManager,
			f.providerRepo,
			f.sessionRepo,
			f.tokenService,
			f.passwordEncrypter,
			f.sessionEncrypter,
			f.refreshTokenDuration,
			f.accessTokenDuration,
		)
	}
	return f.loginUserHandler
}

func (f *AuthHandlerFactory) RedirectGoogleHandler() *command.RedirectGoogleHandler {
	if f.redirectGoogleHandler == nil {
		f.redirectGoogleHandler = command.NewRedirectGoogleHandler(f.oauthProvider, f.stateService)
	}
	return f.redirectGoogleHandler
}

func (f *AuthHandlerFactory) LoginGoogleHandler() *command.LoginGoogleHandler {
	if f.loginGoogleHandler == nil {
		f.loginGoogleHandler = command.NewLoginGoogleHandler(
			f.txManager,
			f.oauthProvider,
			f.userRepo,
			f.providerRepo,
			f.profileRepo,
			f.sessionRepo,
			f.tokenService,
			f.sessionEncrypter,
			f.stateService,
			f.refreshTokenDuration,
			f.accessTokenDuration,
		)
	}
	return f.loginGoogleHandler
}

func (f *AuthHandlerFactory) RefreshTokenHandler() *command.RefreshTokenHandler {
	if f.refreshTokenHandler == nil {
		f.refreshTokenHandler = command.NewRefreshTokenHandler(
			f.txManager,
			f.sessionRepo,
			f.blacklistRepo,
			f.tokenService,
			f.sessionEncrypter,
			f.refreshTokenDuration,
			f.accessTokenDuration,
		)
	}
	return f.refreshTokenHandler
}

func (f *AuthHandlerFactory) LogoutHandler() *command.LogoutHandler {
	if f.logoutHandler == nil {
		f.logoutHandler = command.NewLogoutHandler(f.sessionRepo, f.blacklistRepo, f.sessionEncrypter)
	}
	return f.logoutHandler
}

func (f *AuthHandlerFactory) RefreshTokenDuration() time.Duration {
	return f.refreshTokenDuration
}
