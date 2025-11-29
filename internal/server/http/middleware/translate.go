package http_middleware

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/i18n"
)

func LocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := i18n.DetectLanguage(r)
		ctx := i18n.WithLocale(r.Context(), locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
