package http_handler

import (
	"context"

	myi18n "github.com/brunoibarbosa/url-shortener/internal/i18n"
)

type ErrorDetails interface {
	ToArray() []interface{}
}

type SingleDetail struct {
	data map[string]interface{}
}

func (d SingleDetail) ToArray() []interface{} {
	return []interface{}{d.data}
}

type MultipleDetails struct {
	items []map[string]interface{}
}

func (d MultipleDetails) ToArray() []interface{} {
	result := make([]interface{}, len(d.items))
	for i, item := range d.items {
		result[i] = item
	}
	return result
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	SubCode string `json:"sub_code"`
}

func Detail(ctx context.Context, field, messageID string) ErrorDetails {
	message := myi18n.T(ctx, messageID, nil)

	return SingleDetail{data: map[string]interface{}{
		"field":    field,
		"message":  message,
		"sub_code": messageID,
	}}
}

func ValidationErrors(errors []ValidationError) ErrorDetails {
	items := make([]map[string]interface{}, len(errors))
	for i, err := range errors {
		items[i] = map[string]interface{}{
			"field":    err.Field,
			"message":  err.Message,
			"sub_code": err.SubCode,
		}
	}
	return MultipleDetails{items: items}
}
