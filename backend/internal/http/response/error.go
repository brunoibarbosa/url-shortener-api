package response

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func JSONError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Code:    code,
		},
	})
}
