package validation

import (
	"net/url"
	"regexp"
	"strings"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
)

var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

var invalidCharRegex = regexp.MustCompile(`[^\w\.-]`)

func ValidateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)

	if err != nil {
		return domain.ErrInvalidURLFormat
	}

	if parsedURL.Scheme == "" {
		return domain.ErrMissingURLSchema
	}

	if !allowedSchemes[strings.ToLower(parsedURL.Scheme)] {
		return domain.ErrUnsupportedURLSchema
	}

	if parsedURL.Host == "" {
		return domain.ErrMissingURLHost
	}

	host := parsedURL.Hostname()
	if invalidCharRegex.MatchString(host) {
		return domain.ErrInvalidURLFormat
	}

	if strings.ContainsAny(rawURL, " \t\n\r") {
		return domain.ErrInvalidURLFormat
	}

	return nil
}
