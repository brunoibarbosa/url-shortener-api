package main

import (
	"log"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
)

func listenAndServe(r *http_router.AppRouter, cfg config.AppConfig) {
	addr := cfg.Env.ListenAddress
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("Server running on %s\n", addr)
	http.ListenAndServe(addr, r)
}
