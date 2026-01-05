package container

import (
	"github.com/brunoibarbosa/url-shortener/internal/app/session/query"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
)

type SessionHandlerFactory struct {
	listSessionsRepo session_domain.SessionQueryRepository

	listHandler *query.ListSessionsHandler
}

type SessionFactoryDependencies struct {
	ListSessionsRepo session_domain.SessionQueryRepository
}

func NewSessionHandlerFactory(deps SessionFactoryDependencies) *SessionHandlerFactory {
	return &SessionHandlerFactory{
		listSessionsRepo: deps.ListSessionsRepo,
	}
}

func (f *SessionHandlerFactory) ListSessionsHandler() *query.ListSessionsHandler {
	if f.listHandler == nil {
		f.listHandler = query.NewListSessionsHandler(f.listSessionsRepo)
	}
	return f.listHandler
}
