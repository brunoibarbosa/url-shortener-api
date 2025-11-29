package http_middleware

import (
	"log"
	"net/http"

	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				ctx := r.Context()
				switch err := rec.(type) {
				case error:
					log.Printf("Unexpected error: %v", err)
					http_handler.WriteI18nJSONError(ctx, w, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
				default:
					log.Printf("Unknown panic: %v", rec)
					http_handler.WriteI18nJSONError(ctx, w, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
