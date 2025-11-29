package http

import (
	"net/http"
	"time"
)

type Server struct {
	addr string
}

func NewServer(addr string, router *AppRouter) *http.Server {
	s := &Server{
		addr,
	}

	server := &http.Server{
		Addr:              s.addr,
		Handler:           router,
		IdleTimeout:       time.Minute,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server
}
