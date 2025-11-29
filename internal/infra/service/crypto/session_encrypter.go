package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

type SessionEncrypter struct{}

func NewSessionEncrypter() *SessionEncrypter {
	return &SessionEncrypter{}
}

func (e *SessionEncrypter) HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
