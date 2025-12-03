package http_handler

import (
	"net/http"
)

type HandlerResponse any

type HandlerFunc func(w http.ResponseWriter, r *http.Request) *HTTPError
