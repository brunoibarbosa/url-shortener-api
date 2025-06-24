package main

import (
	"log"
	"net/http"

	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
)

func listenAndServe(r *http_router.AppRouter, cfg AppConfig) {
	addr := cfg.Env.ListenAddress
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("Server running on %s\n", addr)
	http.ListenAndServe(addr, r)
}
