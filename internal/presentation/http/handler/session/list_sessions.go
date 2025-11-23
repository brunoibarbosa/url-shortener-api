package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/session/query"
	"github.com/brunoibarbosa/url-shortener/internal/domain"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type ListSessionsParams struct {
	Limit    uint64                   `json:"limit"`
	Page     uint64                   `json:"page"`
	SortBy   query.ListSessionsSortBy `json:"sortBy"`
	SortKind domain.SortKind          `json:"sortKind"`
}

type Session = struct {
	UserAgent string    `json:"userAgent"`
	IPAddress string    `json:"ipAddress"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ListSessions200Response struct {
	Data  []Session `json:"data"`
	Count uint64    `json:"count"`
	Page  uint64    `json:"page"`
	Limit uint64    `json:"limit"`
}

type ListSessionsHTTPHandler struct {
	qry query.ListSessionsHandler
}

func NewListSessionsHTTPHandler(qry query.ListSessionsHandler) *ListSessionsHTTPHandler {
	return &ListSessionsHTTPHandler{
		qry,
	}
}

func (h *ListSessionsHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	payload, validationErr := validateListSessionsParams(r, ctx)
	if validationErr != nil {
		return nil, validationErr
	}

	params := query.ListSessionsParams{
		Pagination: domain.Pagination{
			Number: payload.Page,
			Size:   payload.Limit,
		},
		SortBy:   payload.SortBy,
		SortKind: payload.SortKind,
	}
	list, count, handleErr := h.qry.Handle(r.Context(), params)
	if handleErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
	}

	sessions := make([]Session, len(list))
	for i, dto := range list {
		sessions[i] = Session{
			UserAgent: dto.UserAgent,
			IPAddress: dto.IPAddress,
			CreatedAt: dto.CreatedAt,
			ExpiresAt: dto.ExpiresAt,
		}
	}

	response := ListSessions200Response{
		Data:  sessions,
		Count: count,
		Page:  payload.Page,
		Limit: payload.Limit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}
func validateListSessionsParams(r *http.Request, ctx context.Context) (ListSessionsParams, *handler.HTTPError) {
	var params ListSessionsParams

	// --------------------------------------------------

	v := r.URL.Query().Get("page")
	if v == "" {
		return ListSessionsParams{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.required_pagination", nil)
	}
	page, parseErr := strconv.ParseUint(v, 10, 64)
	if page == 0 || parseErr != nil {
		return ListSessionsParams{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.invalid_parameter", nil)
	}
	params.Page = page

	// --------------------------------------------------

	v = r.URL.Query().Get("limit")
	if v == "" {
		return ListSessionsParams{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.required_pagination", nil)
	}
	limit, parseErr := strconv.ParseUint(v, 10, 64)
	if limit == 0 || parseErr != nil {
		return ListSessionsParams{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.invalid_parameter", nil)
	}
	params.Limit = limit

	// --------------------------------------------------

	v = r.URL.Query().Get("sortBy")
	var sortBy query.ListSessionsSortBy
	if v != "" {
		switch strings.ToUpper(v) {
		case "USERAGENT":
			sortBy = query.ListSessionsSortByUserAgent
		case "IPADDRESS":
			sortBy = query.ListSessionsSortByIPAddress
		case "CREATEDAT":
			sortBy = query.ListSessionsSortByCreatedAt
		case "EXPIRESAT":
			sortBy = query.ListSessionsSortByExpiresAt
		default:
			return ListSessionsParams{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.invalid_parameter", nil)
		}
	}
	params.SortBy = sortBy

	// --------------------------------------------------

	v = r.URL.Query().Get("sortKind")
	var sortKind domain.SortKind
	if v != "" {
		switch strings.ToUpper(v) {
		case "ASC":
			sortKind = domain.SortAsc
		case "DESC":
			sortKind = domain.SortDesc
		default:
			return ListSessionsParams{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.invalid_parameter", nil)
		}
	}
	params.SortKind = sortKind

	return params, nil
}
