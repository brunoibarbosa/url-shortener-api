package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const SessionIDKey contextKey = "sessionID"

type AuthMiddleware struct {
	Secret []byte
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{Secret: []byte(secret)}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpError := handler.NewI18nHTTPError(r.Context(), http.StatusUnauthorized, errors.CodeUnauthorized, "error.session.missing_access_token", nil)
			handler.WriteJSONError(w, httpError.Status, httpError.Code, httpError.Message, httpError.SubCode)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpError := handler.NewI18nHTTPError(r.Context(), http.StatusUnauthorized, errors.CodeUnauthorized, "error.session.invalid_access_token", nil)
			handler.WriteJSONError(w, httpError.Status, httpError.Code, httpError.Message, httpError.SubCode)
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return m.Secret, nil
		})

		if err != nil || !token.Valid {
			httpError := handler.NewI18nHTTPError(r.Context(), http.StatusUnauthorized, errors.CodeUnauthorized, "error.session.invalid_access_token", nil)
			handler.WriteJSONError(w, httpError.Status, httpError.Code, httpError.Message, httpError.SubCode)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpError := handler.NewI18nHTTPError(r.Context(), http.StatusUnauthorized, errors.CodeUnauthorized, "error.session.invalid_access_token", nil)
			handler.WriteJSONError(w, httpError.Status, httpError.Code, httpError.Message, httpError.SubCode)
			return
		}

		sid, ok := claims["sid"].(string)
		if !ok || sid == "" {
			httpError := handler.NewI18nHTTPError(r.Context(), http.StatusUnauthorized, errors.CodeUnauthorized, "error.session.invalid_access_token", nil)
			handler.WriteJSONError(w, httpError.Status, httpError.Code, httpError.Message, httpError.SubCode)
			return
		}

		ctx := context.WithValue(r.Context(), SessionIDKey, sid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
