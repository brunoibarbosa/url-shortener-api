package http_middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type OptionalAuthMiddleware struct {
	Secret []byte
}

func NewOptionalAuthMiddleware(secret string) *OptionalAuthMiddleware {
	return &OptionalAuthMiddleware{Secret: []byte(secret)}
}

func (m *OptionalAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			next.ServeHTTP(w, r)
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return m.Secret, nil
		})

		if err != nil || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		if sid, ok := claims["sid"].(string); ok && sid != "" {
			ctx = context.WithValue(ctx, SessionIDKey, sid)
		}

		if sub, ok := claims["sub"].(string); ok && sub != "" {
			if userID, err := uuid.Parse(sub); err == nil {
				ctx = context.WithValue(ctx, UserIDKey, userID)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
