package url

import "errors"

var (
	ErrInvalidURLFormat     = errors.New("invalid url")
	ErrMissingURLSchema     = errors.New("url is missing scheme (e.g., http or https)")
	ErrUnsupportedURLSchema = errors.New("unsupported scheme")
	ErrMissingURLHost       = errors.New("url is missing host")
)

type URL struct {
	ShortCode    string
	EncryptedURL string
}
