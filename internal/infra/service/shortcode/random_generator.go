package shortcode

import (
	"crypto/rand"
	"math/big"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
)

type RandomShortCodeGenerator struct{}

func NewRandomShortCodeGenerator() *RandomShortCodeGenerator {
	return &RandomShortCodeGenerator{}
}

func (g *RandomShortCodeGenerator) Generate(length int) (string, error) {
	r := make([]rune, length)
	for i := range r {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(url.ShortCodeCharset))))
		if err != nil {
			return "", err
		}

		r[i] = url.ShortCodeCharset[num.Int64()]
	}

	return string(r), nil
}
