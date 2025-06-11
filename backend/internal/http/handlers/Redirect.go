package handlers

import (
	"net/http"

	"github.com/brunoibarbosa/encurtador-go/internal/config"
	"github.com/brunoibarbosa/encurtador-go/internal/storage"
	crypto "github.com/brunoibarbosa/encurtador-go/pkg/crypto"
)

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortId := r.URL.Path[1:]
	encryptedUrl, ok := storage.GetURL(shortId)

	if !ok {
		http.Error(w, "Esta URL não existe", http.StatusNotFound)
		return
	}

	decryptedUrl := crypto.Decrypt(encryptedUrl, config.Env.SecretKey)
	http.Redirect(w, r, decryptedUrl, http.StatusFound)
}
