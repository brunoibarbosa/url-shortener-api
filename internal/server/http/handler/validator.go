package http_handler

import (
	"context"

	myi18n "github.com/brunoibarbosa/url-shortener/internal/i18n"
)

type ErrorCollector struct {
	ctx            context.Context
	validationErrs []ValidationError
}

func NewErrorCollector(ctx context.Context) *ErrorCollector {
	return &ErrorCollector{
		ctx:            ctx,
		validationErrs: make([]ValidationError, 0),
	}
}

func (v *ErrorCollector) AddError(field, messageID string) {
	message := myi18n.T(v.ctx, messageID, nil)

	err := ValidationError{
		Field:   field,
		Message: message,
		SubCode: messageID,
	}
	v.validationErrs = append(v.validationErrs, err)
}

func (v *ErrorCollector) AddFieldError(field, messageID string) {
	v.AddError(field, messageID)
}

func (v *ErrorCollector) HasErrors() bool {
	return len(v.validationErrs) > 0
}

func (v *ErrorCollector) ToHTTPError(status int, code, messageID string) *HTTPError {
	if len(v.validationErrs) == 0 {
		return nil
	}

	return NewI18nHTTPError(v.ctx, status, code, messageID, ValidationErrors(v.validationErrs))
}
