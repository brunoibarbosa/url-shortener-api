package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorCodeStruct struct {
	InvalidRequest string
	NotFound       string
}

var ErrorCode = ErrorCodeStruct{
	InvalidRequest: "INVALID_REQUEST",
	NotFound:       "NOT_FOUND",
}

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
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
