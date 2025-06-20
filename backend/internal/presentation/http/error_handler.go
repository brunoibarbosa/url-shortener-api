package handler

import (
	"context"
	"encoding/json"
	"net/http"

	myi18n "github.com/brunoibarbosa/url-shortener/internal/i18n"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type HTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, code, message string) *HTTPError {
	return &HTTPError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func NewI18nHTTPError(ctx context.Context, status int, code, messageID string, data map[string]interface{}) *HTTPError {
	return &HTTPError{
		Status:  status,
		Code:    code,
		Message: myi18n.T(ctx, messageID, data),
	}
}

func WriteI18nJSONError(ctx context.Context, w http.ResponseWriter, status int, code string, messageID string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Message: myi18n.T(ctx, messageID, data),
			Code:    code,
		},
	})
}

func WriteJSONError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Code:    code,
		},
	})
}
