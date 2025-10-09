package validation

import (
	"unicode"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

func ValidatePassword(password string) error {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 8 {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasMinLen {
		return domain.ErrPasswordTooShort
	}
	if !hasUpper {
		return domain.ErrPasswordMissingUpper
	}
	if !hasLower {
		return domain.ErrPasswordMissingLower
	}
	if !hasNumber {
		return domain.ErrPasswordMissingDigit
	}
	if !hasSpecial {
		return domain.ErrPasswordMissingSymbol
	}

	return nil
}
