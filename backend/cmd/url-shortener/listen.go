package main

import (
	"log"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/go-chi/chi/v5"
)

func listenAndServe(r *chi.Mux, cfg config.AppConfig) {
	addr := cfg.Env.ListenAddress
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("Server running on %s\n", addr)
	http.ListenAndServe(addr, r)
}
