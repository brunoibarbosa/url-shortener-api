package session

type SessionEncrypter interface {
	HashRefreshToken(token string) string
}
