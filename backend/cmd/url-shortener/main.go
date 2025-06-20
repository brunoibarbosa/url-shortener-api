package main

import (
	"log"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/i18n"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
)

func main() {
	cfg := config.Load()

	if err := i18n.Init(); err != nil {
		log.Fatalf("failed to initialize i18n: %v", err)
	}

	r := http_router.NewRouter(cfg)

	listenAndServe(r, cfg)
}
