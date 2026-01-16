//go:build test

package testhelpers

import (
	"math/rand"
	"strings"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	alphaChars       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphaNumericChar = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	numericChars     = "0123456789"
)

// AlphaString generates a random alphabetic string with length between min and max
func AlphaString(min, max int) string {
	length := min
	if max > min {
		length = min + rng.Intn(max-min+1)
	}
	return randomString(alphaChars, length)
}

// AlphaNumericString generates a random alphanumeric string with length between min and max
func AlphaNumericString(min, max int) string {
	length := min
	if max > min {
		length = min + rng.Intn(max-min+1)
	}
	return randomString(alphaNumericChar, length)
}

// NumericString generates a random numeric string with length between min and max
func NumericString(min, max int) string {
	length := min
	if max > min {
		length = min + rng.Intn(max-min+1)
	}
	return randomString(numericChars, length)
}

// Email generates a random valid email address
func Email() string {
	username := AlphaString(5, 15)
	domain := AlphaString(3, 10)
	tld := []string{"com", "net", "org", "io", "dev"}[rng.Intn(5)]
	return strings.ToLower(username + "@" + domain + "." + tld)
}

// URL generates a random valid URL
func URL() string {
	protocol := []string{"http", "https"}[rng.Intn(2)]
	domain := AlphaString(5, 12)
	tld := []string{"com", "net", "org", "io"}[rng.Intn(4)]
	path := "/" + AlphaString(3, 10)
	return strings.ToLower(protocol + "://" + domain + "." + tld + path)
}

// ShortCode generates a random short code (6 alphanumeric characters)
func ShortCode() string {
	return AlphaNumericString(6, 6)
}

// Int generates a random integer between min and max (inclusive)
func Int(min, max int) int {
	if min == max {
		return min
	}
	return min + rng.Intn(max-min+1)
}

// Bool generates a random boolean
func Bool() bool {
	return rng.Intn(2) == 1
}

// UserAgent generates a random user agent string
func UserAgent() string {
	browsers := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	}
	return browsers[rng.Intn(len(browsers))]
}

// IPAddress generates a random IPv4 address
func IPAddress() string {
	return NumericString(1, 3) + "." +
		NumericString(1, 3) + "." +
		NumericString(1, 3) + "." +
		NumericString(1, 3)
}

func randomString(charset string, length int) string {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rng.Intn(len(charset))])
	}
	return sb.String()
}
