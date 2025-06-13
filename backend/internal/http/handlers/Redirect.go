package handlers

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/config"
	"github.com/brunoibarbosa/url-shortener/internal/http/response"
	"github.com/brunoibarbosa/url-shortener/internal/storage"
	crypto "github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortId := r.URL.Path[1:]
	encryptedUrl, ok := storage.GetURL(shortId)

	if !ok {
		response.JSONError(w, http.StatusBadRequest, response.ErrorCode.NotFound, "The requested URL was not found")
		return
	}

	decryptedUrl := crypto.Decrypt(encryptedUrl, config.Env.SecretKey)
	http.Redirect(w, r, decryptedUrl, http.StatusFound)
}
