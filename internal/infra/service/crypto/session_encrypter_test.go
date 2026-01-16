package crypto_test

import (
	"strings"
	"testing"

	"github.com/brunoibarbosa/url-shortener/internal/infra/service/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNewSessionEncrypter(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	assert.NotNil(t, encrypter)
}

func TestSessionEncrypter_HashRefreshToken_Success(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token := "my-refresh-token-123"

	hash := encrypter.HashRefreshToken(token)

	assert.NotEmpty(t, hash)
	assert.NotEqual(t, token, hash)
}

func TestSessionEncrypter_HashRefreshToken_Deterministic(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token := "deterministic-token"

	hash1 := encrypter.HashRefreshToken(token)
	hash2 := encrypter.HashRefreshToken(token)

	// SHA256 should be deterministic
	assert.Equal(t, hash1, hash2)
}

func TestSessionEncrypter_HashRefreshToken_DifferentTokensProduceDifferentHashes(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token1 := "token-one"
	token2 := "token-two"

	hash1 := encrypter.HashRefreshToken(token1)
	hash2 := encrypter.HashRefreshToken(token2)

	assert.NotEqual(t, hash1, hash2)
}

func TestSessionEncrypter_HashRefreshToken_EmptyToken(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token := ""

	hash := encrypter.HashRefreshToken(token)

	assert.NotEmpty(t, hash)
	// SHA256 of empty string is a known hash
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
}

func TestSessionEncrypter_HashRefreshToken_LongToken(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token := strings.Repeat("a", 10000)

	hash := encrypter.HashRefreshToken(token)

	assert.NotEmpty(t, hash)
	// SHA256 always produces 64 character hex string (256 bits / 4 bits per hex char)
	assert.Len(t, hash, 64)
}

func TestSessionEncrypter_HashRefreshToken_OutputIsHex(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token := "test-token"

	hash := encrypter.HashRefreshToken(token)

	// Verify all characters are valid hex
	for _, c := range hash {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
		assert.True(t, isHex, "Character %c is not valid lowercase hex", c)
	}
}

func TestSessionEncrypter_HashRefreshToken_ConsistentLength(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	tokens := []string{
		"short",
		"medium-length-token",
		"very-long-token-with-lots-of-characters-to-test-sha256-output",
		"",
		"🔐🔑", // Unicode
	}

	for _, token := range tokens {
		t.Run(token, func(t *testing.T) {
			hash := encrypter.HashRefreshToken(token)

			// SHA256 always produces 64 character hex string
			assert.Len(t, hash, 64)
		})
	}
}

func TestSessionEncrypter_HashRefreshToken_SpecialCharacters(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	tokens := []string{
		"token-with-dashes",
		"token_with_underscores",
		"token.with.dots",
		"token with spaces",
		"token\nwith\nnewlines",
		"token\twith\ttabs",
		"пароль", // Cyrillic
		"密码",     // Chinese
		"🔐🔑",     // Emojis
	}

	hashes := make(map[string]bool)

	for _, token := range tokens {
		t.Run(token, func(t *testing.T) {
			hash := encrypter.HashRefreshToken(token)

			assert.NotEmpty(t, hash)
			assert.Len(t, hash, 64)

			// Each different token should produce a different hash
			assert.False(t, hashes[hash], "Hash collision detected for token: %s", token)
			hashes[hash] = true
		})
	}
}

func TestSessionEncrypter_HashRefreshToken_SimilarTokensProduceDifferentHashes(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	// Very similar tokens
	token1 := "token-12345"
	token2 := "token-12346"

	hash1 := encrypter.HashRefreshToken(token1)
	hash2 := encrypter.HashRefreshToken(token2)

	assert.NotEqual(t, hash1, hash2)
	// SHA256 has avalanche effect - small change produces very different hash
}

func TestSessionEncrypter_HashRefreshToken_CaseSensitive(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	tokenLower := "mytoken"
	tokenUpper := "MYTOKEN"

	hash1 := encrypter.HashRefreshToken(tokenLower)
	hash2 := encrypter.HashRefreshToken(tokenUpper)

	assert.NotEqual(t, hash1, hash2)
}

func TestSessionEncrypter_HashRefreshToken_WhitespaceMatters(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token1 := "token"
	token2 := " token"
	token3 := "token "

	hash1 := encrypter.HashRefreshToken(token1)
	hash2 := encrypter.HashRefreshToken(token2)
	hash3 := encrypter.HashRefreshToken(token3)

	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
	assert.NotEqual(t, hash2, hash3)
}

func TestSessionEncrypter_HashRefreshToken_MultipleInstances(t *testing.T) {
	encrypter1 := crypto.NewSessionEncrypter()
	encrypter2 := crypto.NewSessionEncrypter()

	token := "test-token"

	hash1 := encrypter1.HashRefreshToken(token)
	hash2 := encrypter2.HashRefreshToken(token)

	// Different instances should produce same hash for same input
	assert.Equal(t, hash1, hash2)
}

func TestSessionEncrypter_HashRefreshToken_ConcurrentHashing(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	token := "concurrent-token"
	iterations := 100

	results := make(chan string, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			hash := encrypter.HashRefreshToken(token)
			results <- hash
		}()
	}

	// Collect all results
	firstHash := <-results
	for i := 1; i < iterations; i++ {
		hash := <-results
		assert.Equal(t, firstHash, hash, "Concurrent hashing produced different results")
	}
}

func TestSessionEncrypter_HashRefreshToken_KnownVectors(t *testing.T) {
	encrypter := crypto.NewSessionEncrypter()

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			input:    "test",
			expected: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			hash := encrypter.HashRefreshToken(tc.input)
			assert.Equal(t, tc.expected, hash)
		})
	}
}
