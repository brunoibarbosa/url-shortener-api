package shortcode_test

import (
	"testing"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/shortcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRandomShortCodeGenerator(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	assert.NotNil(t, generator)
}

func TestRandomShortCodeGenerator_Generate_Success(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 8

	code, err := generator.Generate(length)

	require.NoError(t, err)
	assert.Len(t, code, length)
}

func TestRandomShortCodeGenerator_Generate_DifferentLengths(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	lengths := []int{4, 6, 8, 10, 12, 16, 32}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			code, err := generator.Generate(length)

			require.NoError(t, err)
			assert.Len(t, code, length)
		})
	}
}

func TestRandomShortCodeGenerator_Generate_ZeroLength(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	code, err := generator.Generate(0)

	require.NoError(t, err)
	assert.Empty(t, code)
}

func TestRandomShortCodeGenerator_Generate_OnlyValidCharacters(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 100
	code, err := generator.Generate(length)

	require.NoError(t, err)

	// Verify all characters are from the charset
	for _, char := range code {
		found := false
		for _, validChar := range url.ShortCodeCharset {
			if char == validChar {
				found = true
				break
			}
		}
		assert.True(t, found, "Character %c is not in charset", char)
	}
}

func TestRandomShortCodeGenerator_Generate_CodesAreRandom(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 8
	iterations := 100
	codes := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		code, err := generator.Generate(length)
		require.NoError(t, err)

		// Each code should be unique (statistically very unlikely to collide)
		assert.False(t, codes[code], "Generated duplicate code: %s", code)
		codes[code] = true
	}

	// We should have generated 100 unique codes
	assert.Len(t, codes, iterations)
}

func TestRandomShortCodeGenerator_Generate_ContainsAlphanumeric(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	// Generate many codes to ensure we get good variety
	length := 1000
	code, err := generator.Generate(length)

	require.NoError(t, err)

	hasLowercase := false
	hasUppercase := false
	hasDigit := false

	for _, char := range code {
		if char >= 'a' && char <= 'z' {
			hasLowercase = true
		}
		if char >= 'A' && char <= 'Z' {
			hasUppercase = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	// With 1000 characters, we should statistically get all types
	assert.True(t, hasLowercase, "Generated code should contain lowercase letters")
	assert.True(t, hasUppercase, "Generated code should contain uppercase letters")
	assert.True(t, hasDigit, "Generated code should contain digits")
}

func TestRandomShortCodeGenerator_Generate_NoSpecialCharacters(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 100
	code, err := generator.Generate(length)

	require.NoError(t, err)

	// Verify no special characters
	for _, char := range code {
		isAlphanumeric := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9')
		assert.True(t, isAlphanumeric, "Character %c should be alphanumeric", char)
	}
}

func TestRandomShortCodeGenerator_Generate_Consistency(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 8

	// Generate multiple codes
	for i := 0; i < 10; i++ {
		code, err := generator.Generate(length)

		require.NoError(t, err)
		assert.Len(t, code, length)
		assert.NotEmpty(t, code)
	}
}

func TestRandomShortCodeGenerator_Generate_LargeLength(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 10000

	code, err := generator.Generate(length)

	require.NoError(t, err)
	assert.Len(t, code, length)
}

func TestRandomShortCodeGenerator_Generate_ConcurrentGeneration(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 8
	iterations := 100

	results := make(chan string, iterations)
	errors := make(chan error, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			code, err := generator.Generate(length)
			if err != nil {
				errors <- err
			} else {
				results <- code
			}
		}()
	}

	// Collect results
	codes := make(map[string]bool)
	for i := 0; i < iterations; i++ {
		select {
		case code := <-results:
			assert.Len(t, code, length)
			codes[code] = true
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Most codes should be unique (allowing for small probability of collision)
	assert.Greater(t, len(codes), iterations*95/100, "Expected at least 95%% unique codes")
}

func TestRandomShortCodeGenerator_Generate_Distribution(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	// Generate many short codes and check character distribution
	iterations := 1000
	length := 1
	charCount := make(map[rune]int)

	for i := 0; i < iterations; i++ {
		code, err := generator.Generate(length)
		require.NoError(t, err)
		charCount[rune(code[0])]++
	}

	// We should have seen a good variety of characters
	// With 1000 iterations and 62 possible characters, we expect some variety
	assert.Greater(t, len(charCount), 30, "Expected reasonable character distribution")
}

func TestRandomShortCodeGenerator_Generate_CharsetConsistency(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	// Generate multiple codes and verify they all use the same charset
	length := 100
	allChars := make(map[rune]bool)

	for i := 0; i < 10; i++ {
		code, err := generator.Generate(length)
		require.NoError(t, err)

		for _, char := range code {
			allChars[char] = true
		}
	}

	// All characters should be from the valid charset
	for char := range allChars {
		found := false
		for _, validChar := range url.ShortCodeCharset {
			if char == validChar {
				found = true
				break
			}
		}
		assert.True(t, found, "Character %c is not in charset", char)
	}
}

func TestRandomShortCodeGenerator_Generate_RepeatedCalls(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 8
	previousCode := ""

	for i := 0; i < 20; i++ {
		code, err := generator.Generate(length)

		require.NoError(t, err)
		assert.Len(t, code, length)

		// Each code should be different from the previous one
		// (statistically extremely likely with 62^8 possibilities)
		if i > 0 {
			assert.NotEqual(t, previousCode, code, "Generated same code twice in a row")
		}
		previousCode = code
	}
}

func TestRandomShortCodeGenerator_Generate_URLSafeCharacters(t *testing.T) {
	generator := shortcode.NewRandomShortCodeGenerator()

	length := 100
	code, err := generator.Generate(length)

	require.NoError(t, err)

	// Verify all characters are URL-safe (no encoding needed)
	urlUnsafeChars := []rune{'/', '?', '#', '&', '=', '+', ' ', '%'}
	for _, char := range code {
		for _, unsafeChar := range urlUnsafeChars {
			assert.NotEqual(t, char, unsafeChar, "Generated code contains URL-unsafe character: %c", char)
		}
	}
}
