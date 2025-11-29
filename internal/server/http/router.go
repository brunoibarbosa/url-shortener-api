package http

import (
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	"github.com/go-chi/chi/v5"
)

type AppRouter struct {
	*chi.Mux
}

func NewRouter() *AppRouter {
	cr := chi.NewRouter()
	return &AppRouter{
		Mux: cr,
	}
}

func (r *AppRouter) Post(pattern string, handlerCb http_handler.HandlerFunc) {
	r.Mux.Post(pattern, http_handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Get(pattern string, handlerCb http_handler.HandlerFunc) {
	r.Mux.Get(pattern, http_handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Put(pattern string, handlerCb http_handler.HandlerFunc) {
	r.Mux.Put(pattern, http_handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Delete(pattern string, handlerCb http_handler.HandlerFunc) {
	r.Mux.Delete(pattern, http_handler.RequestValidator(handlerCb))
}

func (r *AppRouter) Group(fn func(*AppRouter)) {
	r.Mux.Group(func(cr chi.Router) {
		sub := &AppRouter{Mux: cr.(*chi.Mux)}
		fn(sub)
	})
}
