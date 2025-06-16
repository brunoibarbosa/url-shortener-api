package main

import (
	"github.com/brunoibarbosa/url-shortener/internal/config"
	http_router "github.com/brunoibarbosa/url-shortener/internal/infra/presentation/http"
)

func main() {
	cfg := config.Load()

	r := http_router.NewRouter(cfg)

	listenAndServe(r, cfg)
}
