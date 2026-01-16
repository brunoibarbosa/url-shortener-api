package crypto_test

import (
	"testing"

	"github.com/brunoibarbosa/url-shortener/internal/infra/service/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLEncrypter(t *testing.T) {
	secretKey := "12345678901234567890123456789012" // 32 bytes for AES-256
	encrypter := crypto.NewURLEncrypter(secretKey)

	assert.NotNil(t, encrypter)
}

func TestURLEncrypter_Encrypt_Success(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := "https://example.com/very/long/url"

	encrypted, err := encrypter.Encrypt(plainText)

	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plainText, encrypted)
}

func TestURLEncrypter_Decrypt_Success(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := "https://example.com/test"

	encrypted, err := encrypter.Encrypt(plainText)
	require.NoError(t, err)

	decrypted, err := encrypter.Decrypt(encrypted)

	require.NoError(t, err)
	assert.Equal(t, plainText, decrypted)
}

func TestURLEncrypter_EncryptDecrypt_DifferentURLs(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	urls := []string{
		"https://example.com",
		"http://test.com/path?query=value",
		"https://longdomain.com/very/long/path/with/many/segments",
		"https://example.com:8080/api/v1/users/123",
		"https://example.com/path#fragment",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			encrypted, err := encrypter.Encrypt(url)
			require.NoError(t, err)

			decrypted, err := encrypter.Decrypt(encrypted)
			require.NoError(t, err)

			assert.Equal(t, url, decrypted)
		})
	}
}

func TestURLEncrypter_Encrypt_SameTextGeneratesDifferentCipherTexts(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := "https://example.com"

	encrypted1, err := encrypter.Encrypt(plainText)
	require.NoError(t, err)

	encrypted2, err := encrypter.Encrypt(plainText)
	require.NoError(t, err)

	// Due to random IV, same plaintext should generate different ciphertexts
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same plaintext
	decrypted1, err := encrypter.Decrypt(encrypted1)
	require.NoError(t, err)
	assert.Equal(t, plainText, decrypted1)

	decrypted2, err := encrypter.Decrypt(encrypted2)
	require.NoError(t, err)
	assert.Equal(t, plainText, decrypted2)
}

func TestURLEncrypter_Encrypt_EmptyString(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := ""

	encrypted, err := encrypter.Encrypt(plainText)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encrypter.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plainText, decrypted)
}

func TestURLEncrypter_Encrypt_SpecialCharacters(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainTexts := []string{
		"https://example.com/path?q=hello world",
		"https://example.com/пример", // Cyrillic
		"https://example.com/例",      // Chinese
		"https://example.com/🔐",      // Emoji
		"https://example.com/\n\t",
	}

	for _, plainText := range plainTexts {
		t.Run(plainText, func(t *testing.T) {
			encrypted, err := encrypter.Encrypt(plainText)
			require.NoError(t, err)

			decrypted, err := encrypter.Decrypt(encrypted)
			require.NoError(t, err)

			assert.Equal(t, plainText, decrypted)
		})
	}
}

func TestURLEncrypter_Decrypt_InvalidCipherText(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	invalidCipherText := "this-is-not-valid-hex"

	_, err := encrypter.Decrypt(invalidCipherText)

	assert.Error(t, err)
}

func TestURLEncrypter_Decrypt_TooShortCipherText(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	// Valid hex but too short (less than AES block size which is 16 bytes = 32 hex chars)
	shortCipherText := "0123456789abcdef"

	// Should panic due to invalid slice bounds, so we recover from panic
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic
			assert.NotNil(t, r)
		}
	}()

	_, err := encrypter.Decrypt(shortCipherText)

	// If it doesn't panic, it should return an error
	if err == nil {
		t.Error("Expected error or panic for too short cipher text")
	}
}

func TestURLEncrypter_Encrypt_WithInvalidKeySize(t *testing.T) {
	testCases := []struct {
		name      string
		secretKey string
		shouldErr bool
	}{
		{"16 bytes (AES-128)", "1234567890123456", false},
		{"24 bytes (AES-192)", "123456789012345678901234", false},
		{"32 bytes (AES-256)", "12345678901234567890123456789012", false},
		{"Invalid 10 bytes", "1234567890", true},
		{"Invalid 20 bytes", "12345678901234567890", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := crypto.NewURLEncrypter(tc.secretKey)
			plainText := "https://example.com"

			encrypted, err := encrypter.Encrypt(plainText)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, encrypted)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, encrypted)
			}
		})
	}
}

func TestURLEncrypter_Decrypt_WithWrongKey(t *testing.T) {
	secretKey1 := "12345678901234567890123456789012"
	secretKey2 := "98765432109876543210987654321098"

	encrypter1 := crypto.NewURLEncrypter(secretKey1)
	encrypter2 := crypto.NewURLEncrypter(secretKey2)

	plainText := "https://example.com"

	encrypted, err := encrypter1.Encrypt(plainText)
	require.NoError(t, err)

	// Try to decrypt with wrong key
	decrypted, err := encrypter2.Decrypt(encrypted)

	require.NoError(t, err) // CTR mode doesn't fail on wrong key
	assert.NotEqual(t, plainText, decrypted)
	// Decrypted text will be garbage
}

func TestURLEncrypter_LongURL(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	// Very long URL
	longURL := "https://example.com/very/long/path/" + string(make([]byte, 1000))
	for i := 0; i < len(longURL)-35; i++ {
		longURL = longURL[:35+i] + "a" + longURL[36+i:]
	}

	encrypted, err := encrypter.Encrypt(longURL)
	require.NoError(t, err)

	decrypted, err := encrypter.Decrypt(encrypted)
	require.NoError(t, err)

	assert.Equal(t, longURL, decrypted)
}

func TestURLEncrypter_EncryptedOutputIsHex(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := "https://example.com"

	encrypted, err := encrypter.Encrypt(plainText)
	require.NoError(t, err)

	// Verify it's valid hex
	for _, c := range encrypted {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		assert.True(t, isHex, "Character %c is not valid hex", c)
	}
}

func TestURLEncrypter_MultipleEncryptDecrypt(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := "https://example.com"

	// Encrypt multiple times
	for i := 0; i < 10; i++ {
		encrypted, err := encrypter.Encrypt(plainText)
		require.NoError(t, err)

		decrypted, err := encrypter.Decrypt(encrypted)
		require.NoError(t, err)

		assert.Equal(t, plainText, decrypted)
	}
}

func TestURLEncrypter_ConcurrentEncryption(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	encrypter := crypto.NewURLEncrypter(secretKey)

	plainText := "https://example.com"
	iterations := 100

	done := make(chan bool, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			encrypted, err := encrypter.Encrypt(plainText)
			assert.NoError(t, err)

			decrypted, err := encrypter.Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, plainText, decrypted)

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < iterations; i++ {
		<-done
	}
}
