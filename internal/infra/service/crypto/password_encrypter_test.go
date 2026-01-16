package crypto_test

import (
	"testing"

	"github.com/brunoibarbosa/url-shortener/internal/infra/service/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUserPasswordEncrypter(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	assert.NotNil(t, encrypter)
}

func TestUserPasswordEncrypter_HashPassword_Success(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "mysecretpassword123"

	hash, err := encrypter.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestUserPasswordEncrypter_HashPassword_DifferentPasswordsGenerateDifferentHashes(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password1 := "password123"
	password2 := "password456"

	hash1, err := encrypter.HashPassword(password1)
	require.NoError(t, err)

	hash2, err := encrypter.HashPassword(password2)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestUserPasswordEncrypter_HashPassword_SamePasswordGeneratesDifferentHashes(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "samepassword"

	hash1, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	hash2, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	// Bcrypt uses salt, so same password should generate different hashes
	assert.NotEqual(t, hash1, hash2)
}

func TestUserPasswordEncrypter_HashPassword_EmptyPassword(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := ""

	hash, err := encrypter.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestUserPasswordEncrypter_CheckPassword_CorrectPassword(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "correctpassword"

	hash, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	isValid := encrypter.CheckPassword(hash, password)

	assert.True(t, isValid)
}

func TestUserPasswordEncrypter_CheckPassword_IncorrectPassword(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	correctPassword := "correctpassword"
	incorrectPassword := "wrongpassword"

	hash, err := encrypter.HashPassword(correctPassword)
	require.NoError(t, err)

	isValid := encrypter.CheckPassword(hash, incorrectPassword)

	assert.False(t, isValid)
}

func TestUserPasswordEncrypter_CheckPassword_EmptyPasswordAgainstHash(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "mypassword"

	hash, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	isValid := encrypter.CheckPassword(hash, "")

	assert.False(t, isValid)
}

func TestUserPasswordEncrypter_CheckPassword_InvalidHash(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	invalidHash := "not-a-valid-bcrypt-hash"
	password := "somepassword"

	isValid := encrypter.CheckPassword(invalidHash, password)

	assert.False(t, isValid)
}

func TestUserPasswordEncrypter_WithDifferentCosts(t *testing.T) {
	testCases := []struct {
		name string
		cost int
	}{
		{"MinCost", bcrypt.MinCost},
		{"DefaultCost", bcrypt.DefaultCost},
		{"Cost 12", 12},
		{"Cost 14", 14},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypter := crypto.NewUserPasswordEncrypter(tc.cost)
			password := "testpassword123"

			hash, err := encrypter.HashPassword(password)

			require.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Verify the hash can be checked
			isValid := encrypter.CheckPassword(hash, password)
			assert.True(t, isValid)
		})
	}
}

func TestUserPasswordEncrypter_CheckPassword_CaseSensitive(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "MyPassword123"

	hash, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	// Check with different case
	isValid := encrypter.CheckPassword(hash, "mypassword123")
	assert.False(t, isValid)

	// Check with correct case
	isValid = encrypter.CheckPassword(hash, password)
	assert.True(t, isValid)
}

func TestUserPasswordEncrypter_CheckPassword_WhitespaceMatters(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "password"

	hash, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	// Check with leading space
	isValid := encrypter.CheckPassword(hash, " password")
	assert.False(t, isValid)

	// Check with trailing space
	isValid = encrypter.CheckPassword(hash, "password ")
	assert.False(t, isValid)

	// Check exact match
	isValid = encrypter.CheckPassword(hash, password)
	assert.True(t, isValid)
}

func TestUserPasswordEncrypter_LongPassword(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	// Bcrypt has a 72 byte limit, so test with password at the limit
	longPassword := "ThisIsAPasswordExactly72BytesLongToTestBcryptLimitHandlingCorrectly!"

	hash, err := encrypter.HashPassword(longPassword)
	require.NoError(t, err)

	isValid := encrypter.CheckPassword(hash, longPassword)
	assert.True(t, isValid)

	// Test that passwords exceeding 72 bytes will fail
	tooLongPassword := "ThisIsAVeryLongPasswordThatExceedsTheNormalLengthButShouldStillWorkWithBcrypt1234567890"
	_, err = encrypter.HashPassword(tooLongPassword)
	assert.Error(t, err, "bcrypt should error on passwords exceeding 72 bytes")
}

func TestUserPasswordEncrypter_SpecialCharacters(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	passwords := []string{
		"p@ssw0rd!",
		"пароль", // Cyrillic
		"密码",     // Chinese
		"🔐🔑",     // Emojis
		"tab\ttab",
		"newline\nnewline",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := encrypter.HashPassword(password)
			require.NoError(t, err)

			isValid := encrypter.CheckPassword(hash, password)
			assert.True(t, isValid)
		})
	}
}

func TestUserPasswordEncrypter_HashStructure(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "testpassword"

	hash, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	// Bcrypt hash should start with $2a$ or $2b$
	assert.True(t, len(hash) > 10, "Hash should be reasonably long")
	assert.Contains(t, []string{"$2a$", "$2b$"}, hash[:4], "Hash should start with bcrypt prefix")
}

func TestUserPasswordEncrypter_MultipleChecks(t *testing.T) {
	cost := bcrypt.DefaultCost
	encrypter := crypto.NewUserPasswordEncrypter(cost)

	password := "testpassword"

	hash, err := encrypter.HashPassword(password)
	require.NoError(t, err)

	// Check the same hash multiple times
	for i := 0; i < 5; i++ {
		isValid := encrypter.CheckPassword(hash, password)
		assert.True(t, isValid)
	}
}
