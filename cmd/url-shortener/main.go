package main

import (
	"log"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/i18n"
)

func main() {
	cfg := config.Load()

	if err := i18n.Init(); err != nil {
		log.Fatalf("failed to initialize i18n: %v", err)
	}

	r := getRouter(cfg)
	listenAndServe(r, cfg)
}
