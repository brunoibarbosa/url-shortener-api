package url_test

import (
	"testing"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/stretchr/testify/assert"
)

func TestShortCodeCharset(t *testing.T) {
	t.Run("should have correct charset length", func(t *testing.T) {
		// 26 lowercase + 26 uppercase + 10 digits = 62 characters
		assert.Equal(t, 62, len(domain.ShortCodeCharset))
	})

	t.Run("should contain lowercase letters", func(t *testing.T) {
		charset := string(domain.ShortCodeCharset)
		assert.Contains(t, charset, "a")
		assert.Contains(t, charset, "z")
	})

	t.Run("should contain uppercase letters", func(t *testing.T) {
		charset := string(domain.ShortCodeCharset)
		assert.Contains(t, charset, "A")
		assert.Contains(t, charset, "Z")
	})

	t.Run("should contain digits", func(t *testing.T) {
		charset := string(domain.ShortCodeCharset)
		assert.Contains(t, charset, "0")
		assert.Contains(t, charset, "9")
	})

	t.Run("should not contain special characters", func(t *testing.T) {
		charset := string(domain.ShortCodeCharset)
		assert.NotContains(t, charset, "-")
		assert.NotContains(t, charset, "_")
		assert.NotContains(t, charset, "!")
		assert.NotContains(t, charset, "@")
	})
}

func TestShortCodeGenerator_Errors(t *testing.T) {
	t.Run("should have ErrMaxRetries constant", func(t *testing.T) {
		assert.Equal(t, "max retries reached, unable to generate unique short code", domain.ErrMaxRetries.Error())
	})
}
