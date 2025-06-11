package handlers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/brunoibarbosa/encurtador-go/internal/config"
	"github.com/brunoibarbosa/encurtador-go/internal/storage"
	crypto "github.com/brunoibarbosa/encurtador-go/pkg/crypto"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	var payload ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&payload)

	if errors.Is(err, io.EOF) {
		http.Error(w, "Request body must not be empty", http.StatusBadRequest)
		return
	}

	if payload.URL == "" {
		http.Error(w, "'url' field is required in the request body", http.StatusBadRequest)
		return
	}

	if !(strings.HasPrefix(payload.URL, "https://") || strings.HasPrefix(payload.URL, "http://")) {
		http.Error(w, "The 'url' field must start with https:// or http://.", http.StatusBadRequest)
		return
	}

	encryptedUrl := crypto.Encrypt(payload.URL, config.Env.SecretKey)
	shortId := generateShortId()
	storage.SaveURL(shortId, encryptedUrl)

	shortUrl := fmt.Sprintf("http://localhost:8080/%s", shortId)
	resp := ShortenResponse{ShortURL: shortUrl}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func generateShortId() string {
	b := make([]rune, 6)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Fatal(err)
		}

		b[i] = letters[num.Int64()]
	}

	return string(b)
}
