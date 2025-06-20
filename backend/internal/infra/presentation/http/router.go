package http

import (
	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/go-chi/chi/v5"
)

type AppRouter struct {
	*chi.Mux
}

func NewRouter(appConfig config.AppConfig) *AppRouter {
	cr := chi.NewRouter()
	return &AppRouter{
		Mux: cr,
	}
}

func (r *AppRouter) Post(pattern string, handlerCb handler.HandlerFunc) {
	r.Mux.Post(pattern, handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Get(pattern string, handlerCb handler.HandlerFunc) {
	r.Mux.Get(pattern, handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Put(pattern string, handlerCb handler.HandlerFunc) {
	r.Mux.Put(pattern, handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Delete(pattern string, handlerCb handler.HandlerFunc) {
	r.Mux.Delete(pattern, handler.RequestValidator(handlerCb))
}
