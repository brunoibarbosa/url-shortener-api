package url

import (
	"errors"
)

var ShortCodeCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type ShortCodeGenerator interface {
	Generate(length int) (string, error)
}

var ErrMaxRetries = errors.New("max retries reached, unable to generate unique short code")
