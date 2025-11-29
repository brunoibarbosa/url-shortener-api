package http_handler

import (
	"context"
	"encoding/json"
	"net/http"

	myi18n "github.com/brunoibarbosa/url-shortener/internal/i18n"
)

type ErrorDetail struct {
	Code    string        `json:"code"`
	SubCode string        `json:"sub_code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type HTTPError struct {
	Status  int
	Code    string
	SubCode string
	Message string
	Details []interface{}
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewI18nHTTPError(ctx context.Context, status int, code, messageID string, details ErrorDetails) *HTTPError {
	var detailsArray []interface{}
	if details != nil {
		detailsArray = details.ToArray()
	}
	return &HTTPError{
		Status:  status,
		Code:    code,
		SubCode: messageID,
		Message: myi18n.T(ctx, messageID, nil),
		Details: detailsArray,
	}
}

func WriteI18nJSONError(ctx context.Context, w http.ResponseWriter, status int, code string, messageID string, details ErrorDetails) {
	var detailsArray []interface{}
	if details != nil {
		detailsArray = details.ToArray()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Message: myi18n.T(ctx, messageID, nil),
			Code:    code,
			SubCode: messageID,
			Details: detailsArray,
		},
	})
}

func WriteJSONError(w http.ResponseWriter, status int, code, message, subCode string) {
	WriteJSONErrorWithDetails(w, status, code, message, subCode, nil)
}

func WriteJSONErrorWithDetails(w http.ResponseWriter, status int, code, message, subCode string, details []interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Code:    code,
			SubCode: subCode,
			Details: details,
		},
	})
}

func RequestValidator(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := h(w, r)
		if err != nil {
			WriteJSONErrorWithDetails(w, err.Status, err.Code, err.Message, err.SubCode, err.Details)
		}
	}
}
