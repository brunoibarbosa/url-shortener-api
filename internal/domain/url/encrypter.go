package url

type URLEncrypter interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}
