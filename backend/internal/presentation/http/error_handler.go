package handler

import (
	"encoding/json"
	"net/http"
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
