package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/brunoibarbosa/encurtador-go/internal/config"
	"github.com/brunoibarbosa/encurtador-go/internal/http/handlers"
)

func main() {
	config.LoadConfig()
	listen()
}

func listen() {
	http.HandleFunc("/shorten", handlers.ShortenHandler)
	http.HandleFunc("/", handlers.RedirectHandler)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
