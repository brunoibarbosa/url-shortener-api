package user

type UserPasswordEncrypter interface {
	HashPassword(string) (string, error)
	CheckPassword(string, string) bool
}
