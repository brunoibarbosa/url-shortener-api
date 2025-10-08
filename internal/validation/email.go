package validation

import (
	"net/mail"
	"regexp"
	"strings"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

func ValidateEmail(email string) error {
	if isValid := isValidFormat(email); !isValid {
		return domain.ErrInvalidEmailFormat
	}

	return nil
}

func isValidFormat(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	// ParseAddress aceita formas como "Name <local@domain>"
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	// extrai somente o endereço.
	parsed := addr.Address

	// deve conter exatamente um '@'
	parts := strings.Split(parsed, "@")
	if len(parts) != 2 {
		return false
	}
	local := parts[0]
	domain := parts[1]

	// comprimentos máximos por RFC
	if len(local) == 0 || len(local) > 64 {
		return false
	}
	if len(domain) == 0 || len(domain) > 255 {
		return false
	}
	if len(parsed) > 254 {
		return false
	}

	// local-part: não pode começar/terminar com '.' nem ter '..'
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") || strings.Contains(local, "..") {
		return false
	}
	// domínio: não pode ter '..'
	if strings.Contains(domain, "..") {
		return false
	}

	// valida cada label do domínio (comprimento e caracteres permitidos: letters, digits, hyphen)
	labels := strings.Split(domain, ".")
	labelRegexp := regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	for _, lab := range labels {
		if lab == "" {
			return false // label vazia (ex: "ex..com" ou ".com")
		}
		if len(lab) > 63 {
			return false
		}
		if strings.HasPrefix(lab, "-") || strings.HasSuffix(lab, "-") {
			return false
		}
		if !labelRegexp.MatchString(lab) {
			return false
		}
	}

	// checar caracteres do local-part com regexp básica
	localRegexp := regexp.MustCompile(`^[A-Za-z0-9!#\$%&'\*\+/=\?\^_` + "`" + `{|}~\.-]+$`)
	return localRegexp.MatchString(local)

	return true
}
