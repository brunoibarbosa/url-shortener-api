package url

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateShortCode(length int) (string, error) {
	r := make([]rune, length)
	for i := range r {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		r[i] = charset[num.Int64()]
	}

	return string(r), nil
}

var ErrMaxRetries = errors.New("max retries reached, unable to generate unique short code")
