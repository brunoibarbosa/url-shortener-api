package container

import (
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/app/url/query"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
)

type URLHandlerFactory struct {
	persistRepo               domain.URLRepository
	cacheRepo                 domain.URLCacheRepository
	queryRepo                 domain.URLQueryRepository
	encrypter                 domain.URLEncrypter
	shortCodeGenerator        domain.ShortCodeGenerator
	persistExpirationDuration time.Duration
	cacheExpirationDuration   time.Duration

	createHandler *command.CreateShortURLHandler
	deleteHandler *command.DeleteURLHandler
	getHandler    *query.GetOriginalURLHandler
	listHandler   *query.ListUserURLsHandler
}

type URLFactoryDependencies struct {
	PersistRepo               domain.URLRepository
	CacheRepo                 domain.URLCacheRepository
	QueryRepo                 domain.URLQueryRepository
	Encrypter                 domain.URLEncrypter
	ShortCodeGenerator        domain.ShortCodeGenerator
	PersistExpirationDuration time.Duration
	CacheExpirationDuration   time.Duration
}

func NewURLHandlerFactory(deps URLFactoryDependencies) *URLHandlerFactory {
	return &URLHandlerFactory{
		persistRepo:               deps.PersistRepo,
		cacheRepo:                 deps.CacheRepo,
		queryRepo:                 deps.QueryRepo,
		encrypter:                 deps.Encrypter,
		shortCodeGenerator:        deps.ShortCodeGenerator,
		persistExpirationDuration: deps.PersistExpirationDuration,
		cacheExpirationDuration:   deps.CacheExpirationDuration,
	}
}

func (f *URLHandlerFactory) CreateShortURLHandler() *command.CreateShortURLHandler {
	if f.createHandler == nil {
		f.createHandler = command.NewCreateShortURLHandler(
			f.persistRepo,
			f.cacheRepo,
			f.encrypter,
			f.shortCodeGenerator,
			f.persistExpirationDuration,
			f.cacheExpirationDuration,
		)
	}
	return f.createHandler
}

func (f *URLHandlerFactory) GetOriginalURLHandler() *query.GetOriginalURLHandler {
	if f.getHandler == nil {
		f.getHandler = query.NewGetOriginalURLHandler(
			f.persistRepo,
			f.cacheRepo,
			f.encrypter,
			f.cacheExpirationDuration,
		)
	}
	return f.getHandler
}

func (f *URLHandlerFactory) ListUserURLsHandler() *query.ListUserURLsHandler {
	if f.listHandler == nil {
		f.listHandler = query.NewListUserURLsHandler(f.queryRepo)
	}
	return f.listHandler
}

func (f *URLHandlerFactory) DeleteURLHandler() *command.DeleteURLHandler {
	if f.deleteHandler == nil {
		f.deleteHandler = command.NewDeleteURLHandler(f.persistRepo)
	}
	return f.deleteHandler
}
