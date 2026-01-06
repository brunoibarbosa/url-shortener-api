package http_handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/query"
	"github.com/brunoibarbosa/url-shortener/internal/domain"
	url_domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	http_middleware "github.com/brunoibarbosa/url-shortener/internal/server/http/middleware"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/google/uuid"
)

type ListUserURLsParams struct {
	Limit    uint64                    `json:"limit"`
	Page     uint64                    `json:"page"`
	SortBy   url_domain.ListURLsSortBy `json:"sortBy"`
	SortKind domain.SortKind           `json:"sortKind"`
}

type URLItem struct {
	ShortCode string     `json:"shortCode"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

type ListUserURLs200Response struct {
	Data  []URLItem `json:"data"`
	Count uint64    `json:"count"`
	Page  uint64    `json:"page"`
	Limit uint64    `json:"limit"`
}

type ListUserURLsHTTPHandler struct {
	qry *query.ListUserURLsHandler
}

func NewListUserURLsHTTPHandler(qry *query.ListUserURLsHandler) *ListUserURLsHTTPHandler {
	return &ListUserURLsHTTPHandler{qry}
}

func (h *ListUserURLsHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	ctx := r.Context()

	userID, ok := ctx.Value(http_middleware.UserIDKey).(uuid.UUID)
	if !ok {
		return http_handler.NewI18nHTTPError(ctx, http.StatusUnauthorized, errors.CodeUnauthorized, "error.session.missing_access_token", nil)
	}

	payload, validationErr := validateListUserURLsParams(r, ctx)
	if validationErr != nil {
		return validationErr
	}

	params := url_domain.ListURLsParams{
		Pagination: domain.Pagination{
			Number: payload.Page,
			Size:   payload.Limit,
		},
		SortBy:   payload.SortBy,
		SortKind: payload.SortKind,
	}

	list, count, handleErr := h.qry.Handle(ctx, userID, params)
	if handleErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
	}

	urls := make([]URLItem, len(list))
	for i, dto := range list {
		urls[i] = URLItem{
			ShortCode: dto.ShortCode,
			ExpiresAt: dto.ExpiresAt,
			CreatedAt: dto.CreatedAt,
		}
	}

	response := ListUserURLs200Response{
		Data:  urls,
		Count: count,
		Page:  payload.Page,
		Limit: payload.Limit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil
}

func validateListUserURLsParams(r *http.Request, ctx context.Context) (ListUserURLsParams, *http_handler.HTTPError) {
	var params ListUserURLsParams

	ec := http_handler.NewErrorCollector(ctx)

	v := r.URL.Query().Get("page")
	if v == "" {
		ec.AddFieldError("page", "error.details.field_required")
	} else {
		page, parseErr := strconv.ParseUint(v, 10, 64)
		if page == 0 || parseErr != nil {
			ec.AddFieldError("page", "error.details.parameter_must_be_positive")
		}
		params.Page = page
	}

	v = r.URL.Query().Get("limit")
	if v == "" {
		ec.AddFieldError("limit", "error.details.field_required")
	} else {
		limit, parseErr := strconv.ParseUint(v, 10, 64)
		if limit == 0 || parseErr != nil {
			ec.AddFieldError("limit", "error.details.parameter_must_be_positive")
		}
		params.Limit = limit
	}

	if ec.HasErrors() {
		return ListUserURLsParams{}, ec.ToHTTPError(http.StatusBadRequest, errors.CodeValidationError, "error.common.required_pagination")
	}

	v = r.URL.Query().Get("sortBy")
	var sortBy = url_domain.ListURLsSortByNone
	if v != "" {
		switch strings.ToUpper(v) {
		case "CREATEDAT":
			sortBy = url_domain.ListURLsSortByCreatedAt
		case "EXPIRESAT":
			sortBy = url_domain.ListURLsSortByExpiresAt
		default:
			ec.AddFieldError("sortBy", "error.details.parameter_invalid_sort")
		}
	}
	params.SortBy = sortBy

	v = r.URL.Query().Get("sortKind")
	var sortKind = domain.SortNone
	if v != "" {
		switch strings.ToUpper(v) {
		case "ASC":
			sortKind = domain.SortAsc
		case "DESC":
			sortKind = domain.SortDesc
		default:
			ec.AddFieldError("sortKind", "error.details.parameter_invalid_sort")
		}
	}
	params.SortKind = sortKind

	if ec.HasErrors() {
		return ListUserURLsParams{}, ec.ToHTTPError(http.StatusBadRequest, errors.CodeValidationError, "error.validation.failed")
	}

	return params, nil
}
