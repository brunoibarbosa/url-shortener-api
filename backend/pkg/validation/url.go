package validation

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

var invalidCharRegex = regexp.MustCompile(`[^\w\.-]`)

func ValidateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)

	if err != nil {
		return fmt.Errorf("invalid url: %v", err)
	}

	if parsedURL.Scheme == "" {
		return errors.New("url is missing scheme (e.g., http or https)")
	}

	if !allowedSchemes[strings.ToLower(parsedURL.Scheme)] {
		return fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return errors.New("url is missing host")
	}

	host := parsedURL.Hostname()
	if invalidCharRegex.MatchString(host) {
		return fmt.Errorf("host contains invalid characters: %s", host)
	}

	if strings.ContainsAny(rawURL, " \t\n\r") {
		return errors.New("url contains spaces or control characters")
	}

	return nil
}
