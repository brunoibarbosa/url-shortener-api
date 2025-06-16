package middleware

import (
	"log"
	"net/http"

	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				switch err := rec.(type) {
				case *handler.HTTPError:
					handler.WriteJSONError(w, err.Status, err.Code, err.Message)
				case error:
					log.Printf("Unexpected error: %v", err)
					handler.WriteJSONError(w, http.StatusInternalServerError, errors.CodeInternalError, "Internal server error")
				default:
					log.Printf("Unknown panic: %v", rec)
					handler.WriteJSONError(w, http.StatusInternalServerError, errors.CodeInternalError, "Internal server error")
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
