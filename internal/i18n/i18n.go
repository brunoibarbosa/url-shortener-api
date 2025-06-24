package i18n

import (
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translation/*.json
var localeFS embed.FS

type contextKey string

const localeKey contextKey = "locale"

var bundle *i18n.Bundle

func Init() error {
	bundle = i18n.NewBundle(language.English)

	files := []string{
		"translation/en.json",
		"translation/pt.json",
	}
	for _, file := range files {
		if _, err := bundle.LoadMessageFileFS(localeFS, file); err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}
	}

	return nil
}

func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey, locale)
}

func LocalizerFromContext(ctx context.Context) *i18n.Localizer {
	locale, ok := ctx.Value(localeKey).(string)
	if !ok || locale == "" {
		locale = "en"
	}
	return i18n.NewLocalizer(bundle, locale)
}

func DetectLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")
	return lang
}

func T(ctx context.Context, messageID string, templateData map[string]interface{}) string {
	localizer := LocalizerFromContext(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		return messageID
	}
	return msg
}
